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
	openrouter "github.com/revrost/go-openrouter"
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
	if err := printer.Fprint(&buf, ah.FileSet, f); err != nil {
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
		if funcDecl, isFuncDecl := n.(*ast.FuncDecl); isFuncDecl {
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

// BaseStrategy provides common functionality for LLM-based rewriting strategies
type BaseStrategy struct {
	ASTHandler *ASTHandler
	Comment    string
	// Add interface for concrete strategies to implement
	rewriteFunc func(string) (string, error)
}

// getFunctionSource extracts the source code of a function
func (bs *BaseStrategy) getFunctionSource(funcDecl *ast.FuncDecl) (string, error) {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, bs.ASTHandler.FileSet, funcDecl); err != nil {
		return "", fmt.Errorf("failed to extract function source: %w", err)
	}
	return buf.String(), nil
}

// createPrompt creates the prompt for the LLM
func (bs *BaseStrategy) createPrompt(functionSource string) string {
	return fmt.Sprintf(
		`You are a highly skilled Go programming language expert specializing in advanced code obfuscation techniques. Your primary goal is to make code as difficult as possible for humans to analyze and understand, while strictly preserving its original functionality.

Your task: Take the Go function provided below and rewrite it applying **exclusively the Dead Code Insertion technique**. You must add unused variables, meaningless computations, conditions that do not affect the function's main result, or blocks of code that will never execute or whose execution is irrelevant to the core logic. It is crucial that the added code looks plausible but does not alter the semantics or the final outcome of the original function. Avoid obvious insertions like if false {}. Strive to make the insertions varied and integrate them into the code in a way that hinders readability.

CRITICAL REQUIREMENTS:
1.  The function signature must remain EXACTLY the same (name, parameters, return types)
2.  Your response must be valid Go code that can be parsed by the Go parser
3.  Do not change the overall behavior or functionality of the function
4.  STRICTLY preserve the return values - if a function returns values, make sure your refactored version returns values of the same types
5.  If a function returns multiple values, ensure your refactored function returns the same number and type of values
6.  Make sure to maintain error handling patterns
7.  Mandatory start of code with package.
8.  The code must compile and all variables used must be defined.

Example of the transformation:

// --- Example Original Function ---
func calculateSum(a, b int) int {
    // Simple addition function
    return a + b
}
// --- End Example Original Function ---

// --- Example Obfuscated Output (Dead Code Insertion Only) ---
func calculateSum(a, b int) int {
    // Added dead code: unused variables and condition
    tempVar := a*a + b*b - 100
    uselessCounter := 0
    if tempVar > 0 && a > 0 {
        // This block might execute but doesn't affect the sum result
        for i := 0; i < 5; i++ {
            uselessCounter += i * (a - b)
        }
        println("Performed insignificant calculations...") // Side effect not impacting the result
    } else {
         // Alternative dead code path
         _ = tempVar + uselessCounter // Just to use the variables
    }

    // Original function logic is preserved
    result := a + b

    // Another dead code block
    if result != (a + b) {
        // This condition will never be true
        panic("Impossible logic error")
    }

    return result
}
// --- End Example Obfuscated Output ---

Now, please rewrite the following Go function using only Dead Code Insertion:

%s

Return **only** the complete, modified Go function code. Do not include any explanations, comments, introductory text, or markdown formatting. The output must be directly parsable by the standard Go parser (go/parser). Ensure only the code is returned.`,
		functionSource,
	)
}

// cleanResponse cleans and validates the response from LLM
func (bs *BaseStrategy) cleanResponse(response string) (string, error) {
	result := strings.TrimSpace(response)

	// Remove markdown code fences if present
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

	// Basic validation
	if len(result) < 10 {
		return "", fmt.Errorf("received suspiciously short response: %q", result)
	}

	return result, nil
}

// addComment adds a comment to a function declaration
func (bs *BaseStrategy) addComment(funcDecl *ast.FuncDecl, commentText string) {
	comment := &ast.Comment{
		Text:  commentText,
		Slash: funcDecl.Pos(),
	}

	if funcDecl.Doc == nil {
		funcDecl.Doc = &ast.CommentGroup{
			List: []*ast.Comment{comment},
		}
	} else {
		funcDecl.Doc.List = append(funcDecl.Doc.List, comment)
	}
}

// Rewrite implements the RewriteStrategy interface
func (bs *BaseStrategy) Rewrite(f *ast.File) (bool, error) {
	functionsRewritten := false
	functionsEncountered := 0

	// Process each function declaration
	for _, decl := range f.Decls {
		funcDecl, isFuncDecl := decl.(*ast.FuncDecl)
		if !isFuncDecl || funcDecl.Body == nil {
			continue
		}

		functionsEncountered++
		fmt.Printf("Processing function: %s\n", funcDecl.Name.Name)

		// Get the original function source
		functionSource, err := bs.getFunctionSource(funcDecl)
		if err != nil {
			return false, fmt.Errorf("failed to extract function source for %s: %w",
				funcDecl.Name.Name, err)
		}

		// Get the rewritten function source from concrete implementation
		rewrittenSource, err := bs.rewriteFunc(functionSource)
		if err != nil {
			return false, fmt.Errorf("failed to rewrite function %s: %w",
				funcDecl.Name.Name, err)
		}

		// Check if the source actually changed
		if rewrittenSource == functionSource {
			fmt.Printf("LLM didn't make any changes to function %s\n", funcDecl.Name.Name)

			// Add an analyzed-but-unchanged comment
			bs.addComment(funcDecl, bs.Comment+" (analyzed but no changes required)")
			functionsRewritten = true
			continue
		}

		fmt.Printf("Got rewritten source for %s (%d bytes)\n", funcDecl.Name.Name, len(rewrittenSource))
		// Parse the rewritten source code
		rewrittenFile, err := bs.ASTHandler.ParseContent(rewrittenSource)
		if err != nil {
			bs.addComment(funcDecl, fmt.Sprintf("// Failed to parse rewritten function code: %v", err))
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
			bs.addComment(funcDecl, "// Failed to find function in the rewritten code")
			fmt.Printf("Couldn't find function declaration in rewritten code for %s\n", funcDecl.Name.Name)
			continue
		}

		// Replace the function body and add a comment
		funcDecl.Body = rewrittenFunc.Body
		bs.addComment(funcDecl, bs.Comment)

		functionsRewritten = true
		fmt.Printf("Successfully rewrote function: %s\n", funcDecl.Name.Name)
	}

	// Log summary
	fmt.Printf("Rewrite summary: Found %d functions, rewrote %v\n",
		functionsEncountered, functionsRewritten)

	return functionsRewritten, nil
}

// LLMStrategy uses an LLM API to rewrite function bodies
type LLMStrategy struct {
	BaseStrategy
}

// NewLLMStrategy creates a new LLM strategy
func NewLLMStrategy(astHandler *ASTHandler, comment string) *LLMStrategy {
	ls := &LLMStrategy{
		BaseStrategy: BaseStrategy{
			ASTHandler: astHandler,
			Comment:    comment,
		},
	}
	// Set the function to use LLMStrategy's implementation
	ls.rewriteFunc = ls.callGeminiLLM
	return ls
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
	model := client.GenerativeModel("gemini-2.5-flash-preview-04-17")
	model.SetTemperature(0.1)
	model.SetTopK(64)
	model.SetTopP(0.9)
	model.SetMaxOutputTokens(8192)
	model.ResponseMIMEType = "text/plain"

	// Create a chat session
	session := model.StartChat()

	// Prepare the prompt
	prompt := ls.createPrompt(functionSource)

	// Implement retry with exponential backoff
	const maxRetries = 5
	var resp *genai.GenerateContentResponse

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = session.SendMessage(ctx, genai.Text(prompt))

		// If successful, break out of the retry loop
		if err == nil {
			break
		}

		// Handle rate limit errors
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "Too Many Requests") {
			backoffTime := math.Min(math.Pow(2, float64(attempt)), 60)
			waitTime := time.Duration(backoffTime*1000) * time.Millisecond

			fmt.Printf("Rate limited by Gemini API. Attempt %d/%d. Waiting %v before retrying...\n",
				attempt+1, maxRetries, waitTime)

			time.Sleep(waitTime)
			continue
		}

		// For other errors, don't retry
		return "", fmt.Errorf("error sending message to Gemini API: %w", err)
	}

	// Check if we still have an error after all retries
	if err != nil {
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "Too Many Requests") {
			return "", fmt.Errorf("Gemini API rate limit exceeded after %d retries: %w", maxRetries, err)
		}
		return "", fmt.Errorf("error sending message to Gemini API: %w", err)
	}

	// Validate and process response
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil ||
		len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("received empty or invalid response from Gemini API")
	}

	// Build the rewritten code from response parts
	var rewrittenCode strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		rewrittenCode.WriteString(fmt.Sprintf("%v", part))
	}

	return ls.cleanResponse(rewrittenCode.String())
}

// OpenRouterStrategy uses OpenRouter API to rewrite function bodies
type OpenRouterStrategy struct {
	BaseStrategy
}

// NewOpenRouterStrategy creates a new OpenRouter strategy
func NewOpenRouterStrategy(astHandler *ASTHandler, comment string) *OpenRouterStrategy {
	ors := &OpenRouterStrategy{
		BaseStrategy: BaseStrategy{
			ASTHandler: astHandler,
			Comment:    comment,
		},
	}
	// Set the function to use OpenRouterStrategy's implementation
	ors.rewriteFunc = ors.callOpenRouterLLM
	return ors
}

// callOpenRouterLLM makes an API call to OpenRouter LLM to rewrite function code
func (ors *OpenRouterStrategy) callOpenRouterLLM(functionSource string) (string, error) {
	ctx := context.Background()

	// Get API key from environment variable
	apiKey, ok := os.LookupEnv("OPENROUTER_API_KEY")
	if !ok {
		return "", fmt.Errorf("environment variable OPENROUTER_API_KEY not set")
	}

	// Create a new OpenRouter client
	client := openrouter.NewClient(
		apiKey,
		openrouter.WithXTitle("MetamorphLLM"),
		openrouter.WithHTTPReferer("https://github.com/Hekzory/MetamorphLLM"),
	)

	// Prepare the prompt
	prompt := ors.createPrompt(functionSource)

	// Call the OpenRouter API
	resp, err := client.CreateChatCompletion(
		ctx,
		openrouter.ChatCompletionRequest{
			Model: "deepseek/deepseek-chat-v3-0324:free", // Can be configured as needed
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    openrouter.ChatMessageRoleUser,
					Content: openrouter.Content{Text: prompt},
				},
			},
			Temperature: 0.1,
			MaxTokens:   8192,
			TopP:        0.9,
		},
	)

	// Implement retry with exponential backoff
	const maxRetries = 5
	var rewrittenCode string
	attempt := 0

	for attempt < maxRetries {
		if err == nil {
			// Extract the response content
			if len(resp.Choices) > 0 && resp.Choices[0].Message.Content.Text != "" {
				rewrittenCode = resp.Choices[0].Message.Content.Text
				break
			} else {
				err = fmt.Errorf("received empty response from OpenRouter API")
			}
		}

		// Handle rate limit errors
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "Too Many Requests") {
			attempt++
			backoffTime := math.Min(math.Pow(2, float64(attempt)), 60)
			waitTime := time.Duration(backoffTime*1000) * time.Millisecond

			fmt.Printf("Rate limited by OpenRouter API. Attempt %d/%d. Waiting %v before retrying...\n",
				attempt, maxRetries, waitTime)

			time.Sleep(waitTime)

			// Retry the API call
			resp, err = client.CreateChatCompletion(
				ctx,
				openrouter.ChatCompletionRequest{
					Model: "deepseek/deepseek-chat-v3-0324:free",
					Messages: []openrouter.ChatCompletionMessage{
						{
							Role:    openrouter.ChatMessageRoleUser,
							Content: openrouter.Content{Text: prompt},
						},
					},
					Temperature: 0.1,
					MaxTokens:   8192,
					TopP:        0.9,
				},
			)
			continue
		}

		// For other errors, don't retry
		return "", fmt.Errorf("error sending message to OpenRouter API: %w", err)
	}

	// Check if we still have an error after all retries
	if err != nil {
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "Too Many Requests") {
			return "", fmt.Errorf("OpenRouter API rate limit exceeded after %d retries: %w", maxRetries, err)
		}
		return "", fmt.Errorf("error sending message to OpenRouter API: %w", err)
	}

	return ors.cleanResponse(rewrittenCode)
}

// APIType represents the type of API to use for rewriting
type APIType string

const (
	// APITypeGemini represents Google's Gemini API
	APITypeGemini APIType = "gemini"
	// APITypeOpenRouter represents OpenRouter API
	APITypeOpenRouter APIType = "openrouter"
)

// Rewriter orchestrates the code rewriting process
type Rewriter struct {
	FileHandler    *FileHandler
	ASTHandler     *ASTHandler
	Strategy       RewriteStrategy
	DefaultComment string
}

// NewRewriter creates a new Rewriter with default components
func NewRewriter() *Rewriter {
	return &Rewriter{
		FileHandler:    &FileHandler{},
		ASTHandler:     NewASTHandler(),
		Strategy:       NewFunctionCommentStrategy("// This function was rewritten by MetamorphLLM"),
		DefaultComment: "// This function was rewritten by MetamorphLLM",
	}
}

// NewLLMRewriter creates a new Rewriter with LLM strategy
func NewLLMRewriter() *Rewriter {
	return NewLLMRewriterWithAPI(APITypeGemini)
}

// NewLLMRewriterWithAPI creates a new Rewriter with the specified API type
func NewLLMRewriterWithAPI(apiType APIType) *Rewriter {
	astHandler := NewASTHandler()
	var strategy RewriteStrategy
	var commentPrefix string

	switch apiType {
	case APITypeOpenRouter:
		strategy = NewOpenRouterStrategy(astHandler, "// This function was rewritten by OpenRouter LLM")
		commentPrefix = "// This function was rewritten by OpenRouter LLM"
	default: // APITypeGemini or any other case
		strategy = NewLLMStrategy(astHandler, "// This function was rewritten by Gemini LLM")
		commentPrefix = "// This function was rewritten by Gemini LLM"
	}

	return &Rewriter{
		FileHandler:    &FileHandler{},
		ASTHandler:     astHandler,
		Strategy:       strategy,
		DefaultComment: commentPrefix,
	}
}

// SetStrategy changes the rewriting strategy
func (r *Rewriter) SetStrategy(strategy RewriteStrategy) {
	r.Strategy = strategy
}

// RewriteFile reads a file and rewrites its content
func (r *Rewriter) RewriteFile(filePath string) (string, error) {
	content, err := r.FileHandler.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return r.RewriteContent(content)
}

// RewriteContent rewrites Go code using the current strategy
func (r *Rewriter) RewriteContent(content string) (string, error) {
	// Parse the Go source code
	f, err := r.ASTHandler.ParseContent(content)
	if err != nil {
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

	// Add build tag to the rewritten content
	resultWithTag := "// +build rewritten\n\n" + result

	// Check if the content actually changed
	if result == content {
		fmt.Println("WARNING: AST printer output matches original content. Adding success comment anyway.")
		return content + "\n\n// Processed by MetamorphLLM (no changes needed)\n", nil
	}

	return resultWithTag, nil
}

// SaveRewrittenFile saves the content to a file
func (r *Rewriter) SaveRewrittenFile(filePath, content string) error {
	return r.FileHandler.WriteFile(filePath, content)
}

