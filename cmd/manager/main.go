package main

import (
	"flag"
	"fmt"
	"github.com/Hekzory/MetamorphLLM/internal/manager"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Define command-line flags
	rewriterPath := flag.String("rewriter", "rewriter", "Path to the rewriter binary")
	suspiciousPath := flag.String("suspicious", "internal/suspicious/suspicious.go", "Path to the suspicious Go source file to rewrite")
	outputPath := flag.String("output", "", "Path to save the rewritten file (defaults to <input>.rewritten.go)")
	targetBinaryDir := flag.String("target-dir", "cmd/suspicious", "Directory to build the final binary in")
	keepRewritten := flag.Bool("keep", true, "Keep the rewritten files after deployment (default: true)")
	testTimeout := flag.String("timeout", "30s", "Timeout for running tests")
	dryRun := flag.Bool("dry-run", false, "Run without deploying the binary")
	forceRewrite := flag.Bool("force-rewrite", false, "Force rewriting even if rewritten file already exists")
	
	// Parse flags
	flag.Parse()
	
	// Create a new manager
	m := manager.NewManager()
	m.RewriterBinary = *rewriterPath
	m.SuspiciousPath = *suspiciousPath
	m.TargetBinaryDir = *targetBinaryDir
	m.KeepRewritten = *keepRewritten
	m.TestTimeout = *testTimeout
	m.ForceRewrite = *forceRewrite
	
	// Set default output path if not specified
	if *outputPath == "" {
		m.OutputPath = *suspiciousPath + ".rewritten.go"
	} else {
		m.OutputPath = *outputPath
	}
	
	// Print configuration
	fmt.Println("=== MetamorphLLM Manager ===")
	fmt.Println("Configuration:")
	fmt.Printf("  Rewriter binary: %s\n", m.RewriterBinary)
	fmt.Printf("  Suspicious file: %s\n", m.SuspiciousPath)
	fmt.Printf("  Output path: %s\n", m.OutputPath)
	fmt.Printf("  Target binary dir: %s\n", m.TargetBinaryDir)
	fmt.Printf("  Keep rewritten: %v\n", m.KeepRewritten)
	fmt.Printf("  Test timeout: %s\n", m.TestTimeout)
	fmt.Printf("  Dry run: %v\n", *dryRun)
	fmt.Printf("  Force rewrite: %v\n", m.ForceRewrite)
	fmt.Println("===========================")
	
	// Validate that the rewriter binary exists (in PATH or specified location)
	if _, err := exec.LookPath(m.RewriterBinary); err != nil {
		// Check if it's a relative path
		absPath, err := filepath.Abs(m.RewriterBinary)
		if err != nil || !fileExists(absPath) {
			fmt.Fprintf(os.Stderr, "Error: Rewriter binary not found: %s\n", m.RewriterBinary)
			os.Exit(1)
		}
		// Use absolute path
		m.RewriterBinary = absPath
	}
	
	// Validate that the suspicious file exists
	if !fileExists(m.SuspiciousPath) {
		fmt.Fprintf(os.Stderr, "Error: Suspicious file not found: %s\n", m.SuspiciousPath)
		os.Exit(1)
	}
	
	// Run the process
	var err error
	if *dryRun {
		// For dry run, only rewrite and test, but don't deploy
		err = dryRunProcess(m)
	} else {
		// Full process
		err = m.Run()
	}
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// dryRunProcess runs only the rewriting and testing steps without deployment
func dryRunProcess(m *manager.Manager) error {
	fmt.Println("Starting dry run process (no deployment)...")
	
	// Step 1: Run the rewriter
	if err := m.RunRewriter(); err != nil {
		return fmt.Errorf("rewriter step failed: %w", err)
	}
	
	// Step 2: Compile the rewritten code
	if err := m.CompileRewritten(); err != nil {
		return fmt.Errorf("compilation step failed: %w", err)
	}
	
	// Step 3: Run tests
	if err := m.RunTests(); err != nil {
		return fmt.Errorf("testing step failed: %w", err)
	}
	
	fmt.Println("Dry run completed successfully! (No binary was deployed)")
	return nil
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
} 