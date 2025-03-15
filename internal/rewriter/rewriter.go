package rewriter

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
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
		return content + "\n\n// Failed to parse code for rewriting\n", nil
	}
	
	// Apply the rewriting strategy
	rewritten, err := r.Strategy.Rewrite(f)
	if err != nil {
		return content + "\n\n// Error during rewriting\n", nil
	}
	
	// If no changes were made, add a comment to the entire file
	if !rewritten {
		return content + "\n\n// No changes made by the MetamorphLLM\n", nil
	}
	
	// Convert the AST back to a string
	result, err := r.ASTHandler.PrintAST(f)
	if err != nil {
		return content + "\n\n// Failed to print rewritten code\n", nil
	}
	
	return result, nil
}

// SaveRewrittenFile saves the content to a file
func (r *Rewriter) SaveRewrittenFile(filePath, content string) error {
	return r.FileHandler.WriteFile(filePath, content)
} 