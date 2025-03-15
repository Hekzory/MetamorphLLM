package rewriter

import (
	"go/ast"
	"os"
	"strings"
	"testing"
)

// TestFileHandler tests file reading and writing operations
func TestFileHandler(t *testing.T) {
	fh := &FileHandler{}
	
	// Create a temporary file with test content
	content := "Test content"
	tmpfile, err := os.CreateTemp("", "filehandler-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	
	// Test reading file
	readContent, err := fh.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}
	
	if readContent != content {
		t.Errorf("Expected content %q, got %q", content, readContent)
	}
	
	// Test writing file
	outputFile := tmpfile.Name() + ".out"
	defer os.Remove(outputFile)
	
	newContent := "New test content"
	err = fh.WriteFile(outputFile, newContent)
	if err != nil {
		t.Fatalf("Error writing file: %v", err)
	}
	
	// Verify the file was written correctly
	savedContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Error reading saved file: %v", err)
	}
	
	if string(savedContent) != newContent {
		t.Errorf("Expected saved content %q, got %q", newContent, string(savedContent))
	}
}

// TestASTHandler tests parsing and printing AST operations
func TestASTHandler(t *testing.T) {
	ah := NewASTHandler()
	
	// Test parsing valid Go code
	validCode := "package test\n\nfunc example() {\n\tfmt.Println(\"Test\")\n}\n"
	astFile, err := ah.ParseContent(validCode)
	if err != nil {
		t.Fatalf("Failed to parse valid Go code: %v", err)
	}
	
	if astFile == nil {
		t.Fatal("Parsed AST file should not be nil")
	}
	
	// Test parsing invalid Go code
	invalidCode := "this is not valid Go code"
	_, err = ah.ParseContent(invalidCode)
	if err == nil {
		t.Fatal("Parsing invalid Go code should return an error")
	}
	
	// Test printing AST
	printed, err := ah.PrintAST(astFile)
	if err != nil {
		t.Fatalf("Failed to print AST: %v", err)
	}
	
	if !strings.Contains(printed, "func example()") {
		t.Errorf("Printed AST should contain the original function declaration")
	}
}

// TestFunctionCommentStrategy tests the function comment rewriting strategy
func TestFunctionCommentStrategy(t *testing.T) {
	ah := NewASTHandler()
	commentText := "// Test comment"
	strategy := NewFunctionCommentStrategy(commentText)
	
	// Test with code containing functions
	code := `package test

func first() {
	fmt.Println("First function")
}

func second() {
	fmt.Println("Second function")
}`

	astFile, err := ah.ParseContent(code)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}
	
	rewritten, err := strategy.Rewrite(astFile)
	if err != nil {
		t.Fatalf("Rewrite failed: %v", err)
	}
	
	if !rewritten {
		t.Fatal("Expected rewritten to be true for code with functions")
	}
	
	// Test with code without functions
	noFuncCode := "package test\n\nvar x = 10\n"
	noFuncAst, err := ah.ParseContent(noFuncCode)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}
	
	rewritten, err = strategy.Rewrite(noFuncAst)
	if err != nil {
		t.Fatalf("Rewrite failed: %v", err)
	}
	
	if rewritten {
		t.Fatal("Expected rewritten to be false for code without functions")
	}
}

// TestRewriteContent verifies that content has functions rewritten
func TestRewriteContent(t *testing.T) {
	r := NewRewriter()
	
	// Test case with a function
	original := "package example\n\nfunc hello() {\n\tfmt.Println(\"Hello, world!\")\n}\n"
	rewritten, err := r.RewriteContent(original)
	if err != nil {
		t.Fatalf("Error rewriting content: %v", err)
	}
	
	// Check that the result is different from the original
	if rewritten == original {
		t.Error("Rewritten content should be different from the original")
	}
	
	// Check that it contains the expected function comment
	if !strings.Contains(rewritten, "This function was rewritten by MetamorphLLM") {
		t.Error("Rewritten content should contain the function rewrite comment")
	}
	
	// Test case with no functions
	noFuncOriginal := "package example\n\nvar x = 10\n"
	noFuncRewritten, err := r.RewriteContent(noFuncOriginal)
	if err != nil {
		t.Fatalf("Error rewriting content: %v", err)
	}
	
	// Check that it contains the fallback comment for no functions
	if !strings.Contains(noFuncRewritten, "No changes made") {
		t.Error("When no functions are present, should contain the 'no changes made' comment")
	}
	
	// Test case with invalid Go code
	invalidOriginal := "this is not valid Go code"
	invalidRewritten, err := r.RewriteContent(invalidOriginal)
	if err != nil {
		t.Fatalf("Error rewriting content: %v", err)
	}
	
	// Check that it contains the fallback comment for parse errors
	if !strings.Contains(invalidRewritten, "Failed to parse code") {
		t.Error("When given invalid Go code, should contain the parse failure comment")
	}
}

// TestRewriteFile tests file reading and rewriting with functions
func TestRewriteFile(t *testing.T) {
	r := NewRewriter()
	
	// Create a temporary file with test content that includes a function
	content := "package test\n\nfunc example() {\n\tfmt.Println(\"Test\")\n}\n"
	tmpfile, err := os.CreateTemp("", "rewriter-test-*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	
	// Rewrite the file
	rewritten, err := r.RewriteFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Error rewriting file: %v", err)
	}
	
	// Check that the result contains the function rewrite comment
	if !strings.Contains(rewritten, "This function was rewritten by MetamorphLLM") {
		t.Error("Rewritten file should contain the function rewrite comment")
	}
	
	// Test saving the rewritten content
	outputFile := tmpfile.Name() + ".out"
	defer os.Remove(outputFile)
	
	err = r.SaveRewrittenFile(outputFile, rewritten)
	if err != nil {
		t.Fatalf("Error saving rewritten file: %v", err)
	}
	
	// Verify the file was written correctly
	savedContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Error reading saved file: %v", err)
	}
	
	if string(savedContent) != rewritten {
		t.Error("Saved file content does not match rewritten content")
	}
}

// TestSetStrategy tests changing rewriting strategies
func TestSetStrategy(t *testing.T) {
	r := NewRewriter()
	
	// Create a custom strategy
	customComment := "// Custom strategy comment"
	customStrategy := NewFunctionCommentStrategy(customComment)
	
	// Set the custom strategy
	r.SetStrategy(customStrategy)
	
	// Test with the custom strategy
	code := "package test\n\nfunc example() {\n\tfmt.Println(\"Test\")\n}\n"
	rewritten, err := r.RewriteContent(code)
	if err != nil {
		t.Fatalf("Error rewriting content: %v", err)
	}
	
	// Check that it uses the custom comment
	if !strings.Contains(rewritten, customComment) {
		t.Errorf("Rewritten content should contain the custom comment")
	}
	
	// Check that it doesn't contain the default comment
	if strings.Contains(rewritten, r.DefaultComment) {
		t.Errorf("Rewritten content should not contain the default comment")
	}
}

// TestMultipleFunctions ensures that all functions in a file are rewritten
func TestMultipleFunctions(t *testing.T) {
	r := NewRewriter()
	
	// Create a file with multiple functions
	original := `package test

func first() {
	fmt.Println("First function")
}

func second() {
	fmt.Println("Second function")
}

func third() {
	fmt.Println("Third function")
}`

	rewritten, err := r.RewriteContent(original)
	if err != nil {
		t.Fatalf("Error rewriting content: %v", err)
	}
	
	// Count occurrences of the rewrite comment
	count := strings.Count(rewritten, "This function was rewritten by MetamorphLLM")
	
	// Should have rewritten all three functions
	if count != 3 {
		t.Errorf("Expected 3 functions to be rewritten, got %d", count)
	}
}

// TestMockStrategy tests creating a custom mock strategy
func TestMockStrategy(t *testing.T) {
	// Create a mock strategy that implements the RewriteStrategy interface
	mockStrategy := &MockStrategy{shouldRewrite: true}
	
	r := NewRewriter()
	r.SetStrategy(mockStrategy)
	
	code := "package test\n\nfunc example() {}\n"
	_, err := r.RewriteContent(code)
	if err != nil {
		t.Fatalf("Error rewriting content: %v", err)
	}
	
	if !mockStrategy.rewriteCalled {
		t.Error("Strategy's Rewrite method should have been called")
	}
}

// MockStrategy is a test implementation of RewriteStrategy
type MockStrategy struct {
	rewriteCalled bool
	shouldRewrite bool
}

// Rewrite implements the RewriteStrategy interface
func (ms *MockStrategy) Rewrite(f *ast.File) (bool, error) {
	ms.rewriteCalled = true
	return ms.shouldRewrite, nil
} 