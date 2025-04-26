package main

import (
	"flag"
	"fmt"
	"github.com/Hekzory/MetamorphLLM/internal/rewriter"
	"os"
)

func main() {
	// Define command-line flags
	inputFile := flag.String("input", "", "Path to the Go file to rewrite")
	outputFile := flag.String("output", "", "Path to save the rewritten file (defaults to <input>.rewritten.go)")
	apiFlag := flag.String("api", "openrouter", "API to use for rewriting: 'gemini' or 'openrouter'")
	
	// Parse flags
	flag.Parse()
	
	// Determine which API to use
	var apiType rewriter.APIType
	switch *apiFlag {
	case "openrouter":
		apiType = rewriter.APITypeOpenRouter
		fmt.Println("Using OpenRouter API for rewriting")
	default:
		apiType = rewriter.APITypeGemini
		fmt.Println("Using Gemini API for rewriting")
	}
	
	// Create a new rewriter with the specified API
	r := rewriter.NewLLMRewriterWithAPI(apiType)
	
	// Handle non-flag arguments as input files
	if flag.NArg() > 0 && *inputFile == "" {
		*inputFile = flag.Arg(0)
	}
	
	// Validate input
	if *inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: No input file specified")
		fmt.Fprintln(os.Stderr, "Usage: rewriter [options] -input <file.go>")
		flag.PrintDefaults()
		os.Exit(1)
	}
	
	// Set default output file if not specified
	if *outputFile == "" {
		*outputFile = *inputFile + ".rewritten.go"
	}
	
	// Perform the rewriting
	fmt.Printf("Rewriting %s to %s...\n", *inputFile, *outputFile)
	
	// Rewrite the file
	rewritten, err := r.RewriteFile(*inputFile)
	if err != nil {
		fmt.Printf("Error rewriting file: %v\n", err)
		os.Exit(1)
	}
	
	// Save the rewritten content
	err = r.SaveRewrittenFile(*outputFile, rewritten)
	if err != nil {
		fmt.Printf("Error saving rewritten file: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Rewriting completed successfully!")
} 