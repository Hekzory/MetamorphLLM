package rewriter

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// TestRewriteContent verifies that content has the comment added
func TestRewriteContent(t *testing.T) {
	r := NewRewriter()
	
	original := "package example\n\nfunc hello() {\n\tfmt.Println(\"Hello, world!\")\n}\n"
	rewritten := r.RewriteContent(original)
	
	// Check that the result is different from the original
	if rewritten == original {
		t.Error("Rewritten content should be different from the original")
	}
	
	// Check that it contains the expected comment
	if !strings.Contains(rewritten, "// This code would be rewritten") {
		t.Error("Rewritten content should contain the placeholder comment")
	}
}

// TestRewriteFile tests file reading and rewriting
func TestRewriteFile(t *testing.T) {
	r := NewRewriter()
	
	// Create a temporary file with test content
	content := "package test\n\nfunc example() {\n\tfmt.Println(\"Test\")\n}\n"
	tmpfile, err := ioutil.TempFile("", "rewriter-test-*.go")
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
	
	// Check that the result contains the placeholder comment
	if !strings.Contains(rewritten, "// This code would be rewritten") {
		t.Error("Rewritten file should contain the placeholder comment")
	}
	
	// Test saving the rewritten content
	outputFile := tmpfile.Name() + ".out"
	defer os.Remove(outputFile)
	
	err = r.SaveRewrittenFile(outputFile, rewritten)
	if err != nil {
		t.Fatalf("Error saving rewritten file: %v", err)
	}
	
	// Verify the file was written correctly
	savedContent, err := ioutil.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Error reading saved file: %v", err)
	}
	
	if string(savedContent) != rewritten {
		t.Error("Saved file content does not match rewritten content")
	}
} 