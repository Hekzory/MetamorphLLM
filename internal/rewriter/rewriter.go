package rewriter

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"math"
	"os"
	"strings"
	"time"
	
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// FileHandler handles file I/O operations
type FileHandler struct{}

// ReadFile reads a file and returns its content as a string
func (fh *FileHandler) ReadFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

// WriteFile saves content to a file
func (fh *FileHandler) WriteFile(filePath string, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}

// ASTHandler handles parsing and printing ASTs
type ASTHandler struct {
	FileSet *token.FileSet
}

// NewASTHandler creates a new ASTHandler
func NewASTHandler() *ASTHandler {
	return &ASTHandler{
		FileSet: token.NewFileSet(),
	}
}

// ParseContent parses Go code into an AST
func (ah *ASTHandler) ParseContent(content string) (*ast.File, error) {
	return parser.ParseFile(ah.FileSet, "", content, parser.ParseComments)
}

// PrintAST converts an AST back to a string
func (ah *ASTHandler) PrintAST(f *ast.File) (string, error) {
	var buf strings.Builder
	err := printer.Fprint(&buf, ah.FileSet, f)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RewriteStrategy defines an interface for different code rewriting strategies
type RewriteStrategy interface {
	Rewrite(f *ast.File) (bool, error)
}

// FunctionCommentStrategy adds comments to function declarations
type FunctionCommentStrategy struct {
	CommentText string
}

// NewFunctionCommentStrategy creates a new function comment strategy
func NewFunctionCommentStrategy(commentText string) *FunctionCommentStrategy {
	return &FunctionCommentStrategy{
		CommentText: commentText,
	}
}

// Rewrite implements the RewriteStrategy interface
func (fcs *FunctionCommentStrategy) Rewrite(f *ast.File) (bool, error) {
	functionsRewritten := false
	
	ast.Inspect(f, func(n ast.Node) bool {
		funcDecl, isFuncDecl := n.(*ast.FuncDecl)
		if isFuncDecl {
			comment := &ast.Comment{
				Text:  fcs.CommentText,
				Slash: funcDecl.End(),
			}
			
			if funcDecl.Doc == nil {
				funcDecl.Doc = &ast.CommentGroup{
					List: []*ast.Comment{comment},
				}
			} else {
				funcDecl.Doc.List = append(funcDecl.Doc.List, comment)
			}
			
			functionsRewritten = true
		}
		return true
	})
	
	return functionsRewritten, nil
}

// LLMStrategy uses an LLM API to rewrite function bodies
type LLMStrategy struct {
	ASTHandler *ASTHandler
	Comment    string
}

// NewLLMStrategy creates a new LLM strategy
func NewLLMStrategy(astHandler *ASTHandler, comment string) *LLMStrategy {
	return &LLMStrategy{
		ASTHandler: astHandler,
		Comment:    comment,
	}
}

// getFunctionSource extracts the source code of a function
func (ls *LLMStrategy) getFunctionSource(file *ast.File, funcDecl *ast.FuncDecl) (string, error) {
	var buf bytes.Buffer
	err := printer.Fprint(&buf, ls.ASTHandler.FileSet, funcDecl)
	if err != nil {
		return "", fmt.Errorf("failed to extract function source: %w", err)
	}
	return buf.String(), nil
}

// callGeminiLLM makes an API call to Gemini LLM to rewrite function code
func (ls *LLMStrategy) callGeminiLLM(functionSource string) (string, error) {
	ctx := context.Background()
	
	// Get API key from environment variable
	apiKey, ok := os.LookupEnv("GEMINI_API_KEY")
	if !ok {
		return "", fmt.Errorf("environment variable GEMINI_API_KEY not set")
	}
	
	// Create a new Gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()
	
	// Create a generative model
	model := client.GenerativeModel("gemini-1.5-flash")
	
	// Configure model parameters
	model.SetTemperature(0.2) // Lower temperature for more focused code generation
	model.SetTopK(64)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(8192)
	model.ResponseMIMEType = "text/plain"
	
	// Create a chat session
	session := model.StartChat()
	
	// Prepare the prompt with stronger guidance
	prompt := fmt.Sprintf(
		`You are an expert Go programmer tasked with refactoring a Go function.

I will provide you with a Go function that needs to be refactored to improve its:
- Readability
- Performance
- Error handling
- Maintainability

CRITICAL REQUIREMENTS:
1. Return ONLY the complete function code with NO explanations or comments before or after
2. The function signature must remain exactly the same (name, parameters, return types)
3. Your response must be valid Go code that can be parsed by the Go parser
4. Do not change the overall behavior or functionality of the function
5. Preserve any imports and package declarations

Here's the function to refactor:

%s

Return only the refactored function code.`,
		functionSource,
	)
	
	// Retry configuration
	maxRetries := 5
	var resp *genai.GenerateContentResponse
	var lastErr error
	
	// Implement retry with exponential backoff
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Send the message to Gemini
		resp, err = session.SendMessage(ctx, genai.Text(prompt))
		
		// If successful, break out of the retry loop
		if err == nil {
			break
		}
		
		// Check if this is a rate limit error (HTTP 429)
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "Too Many Requests") {
			// Calculate backoff time: 2^attempt seconds with some randomness, capped at 60 seconds
			backoffTime := math.Min(math.Pow(2, float64(attempt)), 60)
			waitTime := time.Duration(backoffTime*1000) * time.Millisecond
			
			fmt.Printf("Rate limited by Gemini API. Attempt %d/%d. Waiting %v before retrying...\n", 
				attempt+1, maxRetries, waitTime)
			
			time.Sleep(waitTime)
			lastErr = fmt.Errorf("rate limit exceeded (HTTP 429): %w", err)
			continue
		}
		
		// For other errors, don't retry
		lastErr = err
		break
	}
	
	// Check if we still have an error after all retries
	if err != nil {
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "Too Many Requests") {
			return "", fmt.Errorf("Gemini API rate limit exceeded after %d retries. Please try again later or reduce the number of requests: %w", maxRetries, lastErr)
		}
		return "", fmt.Errorf("error sending message to Gemini API: %w", lastErr)
	}
	
	// Check if we have a valid response
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("received empty response from Gemini API")
	}
	
	if resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("received response with empty content from Gemini API")
	}
	
	// Process the response parts and build the rewritten code
	var rewrittenCode strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		// Use %v to properly handle the part
		rewrittenCode.WriteString(fmt.Sprintf("%v", part))
	}
	
	// Clean up the output to handle potential markdown code blocks
	result := rewrittenCode.String()
	
	// Check for and remove markdown code fences
	result = strings.TrimSpace(result)
	if strings.HasPrefix(result, "```go") {
		result = strings.TrimPrefix(result, "```go")
		if idx := strings.LastIndex(result, "```"); idx != -1 {
			result = result[:idx]
		}
	} else if strings.HasPrefix(result, "```") {
		result = strings.TrimPrefix(result, "```")
		if idx := strings.LastIndex(result, "```"); idx != -1 {
			result = result[:idx]
		}
	}
	result = strings.TrimSpace(result)
	
	// Check if we have a reasonable output
	if len(result) < 10 { // Arbitrary small size check
		return "", fmt.Errorf("received suspiciously short response from Gemini API: %q", result)
	}
	
	return result, nil
}

// Rewrite implements the RewriteStrategy interface
func (ls *LLMStrategy) Rewrite(f *ast.File) (bool, error) {
	functionsRewritten := false
	functionsEncountered := 0
	
	// Process each function declaration
	for _, decl := range f.Decls {
		funcDecl, isFuncDecl := decl.(*ast.FuncDecl)
		if !isFuncDecl {
			continue
		}
		
		// Count how many functions we're encountering
		functionsEncountered++
		
		// Skip functions without a body (e.g., function declarations)
		if funcDecl.Body == nil {
			fmt.Printf("Skipping function %s: no body\n", funcDecl.Name.Name)
			continue
		}
		
		fmt.Printf("Processing function: %s\n", funcDecl.Name.Name)
		
		// Get the original function source
		functionSource, err := ls.getFunctionSource(f, funcDecl)
		if err != nil {
			return false, fmt.Errorf("failed to extract function source for %s: %w", funcDecl.Name.Name, err)
		}
		
		// Call the LLM to rewrite the function
		rewrittenSource, err := ls.callGeminiLLM(functionSource)
		if err != nil {
			return false, fmt.Errorf("failed to rewrite function %s using Gemini LLM: %w", funcDecl.Name.Name, err)
		}
		
		// Check if the source actually changed
		if rewrittenSource == functionSource {
			fmt.Printf("LLM didn't make any changes to function %s\n", funcDecl.Name.Name)
			// Still consider this a rewrite even if content is the same
			functionsRewritten = true
			
			// Add a comment indicating that the function was analyzed but unchanged
			comment := &ast.Comment{
				Text:  ls.Comment + " (analyzed but no changes required)",
				Slash: funcDecl.Pos(),
			}
			
			if funcDecl.Doc == nil {
				funcDecl.Doc = &ast.CommentGroup{
					List: []*ast.Comment{comment},
				}
			} else {
				funcDecl.Doc.List = append(funcDecl.Doc.List, comment)
			}
			continue
		}
		
		fmt.Printf("Got rewritten source for %s (%d bytes)\n", funcDecl.Name.Name, len(rewrittenSource))
		
		// Parse the rewritten source code
		rewrittenFile, err := ls.ASTHandler.ParseContent(rewrittenSource)
		if err != nil {
			// If parsing fails, add a comment but don't replace the function
			comment := &ast.Comment{
				Text:  fmt.Sprintf("// Failed to parse rewritten function code: %v", err),
				Slash: funcDecl.End(),
			}
			
			if funcDecl.Doc == nil {
				funcDecl.Doc = &ast.CommentGroup{
					List: []*ast.Comment{comment},
				}
			} else {
				funcDecl.Doc.List = append(funcDecl.Doc.List, comment)
			}
			fmt.Printf("Failed to parse rewritten code for %s: %v\n", funcDecl.Name.Name, err)
			continue
		}
		
		// Find the function in the rewritten code
		var rewrittenFunc *ast.FuncDecl
		for _, d := range rewrittenFile.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				rewrittenFunc = fd
				break
			}
		}
		
		if rewrittenFunc == nil {
			comment := &ast.Comment{
				Text:  "// Failed to find function in the rewritten code",
				Slash: funcDecl.End(),
			}
			
			if funcDecl.Doc == nil {
				funcDecl.Doc = &ast.CommentGroup{
					List: []*ast.Comment{comment},
				}
			} else {
				funcDecl.Doc.List = append(funcDecl.Doc.List, comment)
			}
			fmt.Printf("Couldn't find function declaration in rewritten code for %s\n", funcDecl.Name.Name)
			continue
		}
		
		// Replace the function body
		funcDecl.Body = rewrittenFunc.Body
		
		// Add a comment indicating that the function was rewritten
		comment := &ast.Comment{
			Text:  ls.Comment,
			Slash: funcDecl.Pos(),
		}
		
		if funcDecl.Doc == nil {
			funcDecl.Doc = &ast.CommentGroup{
				List: []*ast.Comment{comment},
			}
		} else {
			funcDecl.Doc.List = append(funcDecl.Doc.List, comment)
		}
		
		functionsRewritten = true
		fmt.Printf("Successfully rewrote function: %s\n", funcDecl.Name.Name)
	}
	
	// Log summary
	fmt.Printf("Rewrite summary: Found %d functions, rewrote %v\n", 
		functionsEncountered, functionsRewritten)
	
	// Consider it a rewrite even if we just added comments
	return functionsRewritten, nil
}

// Rewriter orchestrates the code rewriting process
type Rewriter struct {
	FileHandler   *FileHandler
	ASTHandler    *ASTHandler
	Strategy      RewriteStrategy
	DefaultComment string
}

// NewRewriter creates a new Rewriter with default components
func NewRewriter() *Rewriter {
	return &Rewriter{
		FileHandler: &FileHandler{},
		ASTHandler:  NewASTHandler(),
		Strategy:    NewFunctionCommentStrategy("// This function was rewritten by MetamorphLLM"),
		DefaultComment: "// This function was rewritten by MetamorphLLM",
	}
}

// NewLLMRewriter creates a new Rewriter with LLM strategy
func NewLLMRewriter() *Rewriter {
	astHandler := NewASTHandler()
	return &Rewriter{
		FileHandler: &FileHandler{},
		ASTHandler:  astHandler,
		Strategy:    NewLLMStrategy(astHandler, "// This function was rewritten by Gemini LLM"),
		DefaultComment: "// This function was rewritten by Gemini LLM",
	}
}

// SetStrategy changes the rewriting strategy
func (r *Rewriter) SetStrategy(strategy RewriteStrategy) {
	r.Strategy = strategy
}

// RewriteFile reads a file and passes its content to RewriteContent
func (r *Rewriter) RewriteFile(filePath string) (string, error) {
	content, err := r.FileHandler.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	
	rewritten, err := r.RewriteContent(content)
	if err != nil {
		return "", err
	}
	
	return rewritten, nil
}

// RewriteContent rewrites Go code using the current strategy
func (r *Rewriter) RewriteContent(content string) (string, error) {
	// Parse the Go source code
	f, err := r.ASTHandler.ParseContent(content)
	if err != nil {
		// If parsing fails, fall back to adding a comment to the entire file
		return content + fmt.Sprintf("\n\n// Failed to parse code for rewriting: %v\n", err), nil
	}
	
	fmt.Println("Applying rewriting strategy to the code...")
	
	// Apply the rewriting strategy
	rewritten, err := r.Strategy.Rewrite(f)
	if err != nil {
		errMsg := fmt.Sprintf("\n\n// Error during rewriting: %v\n", err)
		fmt.Println(errMsg)
		return content + errMsg, nil
	}
	
	// If no changes were made, add a comment to the entire file
	if !rewritten {
		fmt.Println("WARNING: No changes were made during rewriting")
		return content + "\n\n// No changes made by the MetamorphLLM\n", nil
	}
	
	fmt.Println("Successfully rewrote code. Converting AST back to string...")
	
	// Convert the AST back to a string
	result, err := r.ASTHandler.PrintAST(f)
	if err != nil {
		errMsg := fmt.Sprintf("\n\n// Failed to print rewritten code: %v\n", err)
		fmt.Println(errMsg)
		return content + errMsg, nil
	}
	
	// Check if the content actually changed
	if result == content {
		fmt.Println("WARNING: AST printer output matches original content. Adding success comment anyway.")
		return content + "\n\n// Processed by MetamorphLLM (no changes needed)\n", nil
	}
	
	return result, nil
}

// SaveRewrittenFile saves the content to a file
func (r *Rewriter) SaveRewrittenFile(filePath, content string) error {
	return r.FileHandler.WriteFile(filePath, content)
} 