package manager

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Hekzory/MetamorphLLM/internal/metrics"
)

// Manager handles the automated process of rewriting code, testing, and deploying
type Manager struct {
	RewriterBinary  string
	SuspiciousPath  string // Path to the suspicious source file (e.g., internal/suspicious/suspicious.go)
	OutputPath      string // Path for the rewritten source file
	TargetBinaryDir string // Directory where the final binary should be built (e.g., cmd/suspicious)
	TestTimeout     string
	KeepRewritten   bool
	ForceRewrite    bool
}

// NewManager creates a new Manager instance with default values
func NewManager() *Manager {
	return &Manager{
		RewriterBinary:  "rewriter",
		SuspiciousPath:  "internal/suspicious/suspicious.go",              // Default to the actual logic file
		OutputPath:      "internal/suspicious/suspicious.go.rewritten.go", // Default rewritten output path
		TargetBinaryDir: "cmd/suspicious",                                 // Default directory for the final binary
		TestTimeout:     "30s",
		KeepRewritten:   true, // Default to keeping rewritten files
		ForceRewrite:    false,
	}
}

// RunRewriter executes the rewriter binary to generate rewritten code
func (m *Manager) RunRewriter() error {
	fmt.Println("Running rewriter...")

	// Check if the rewritten file already exists
	if !m.ForceRewrite {
		if _, err := os.Stat(m.OutputPath); err == nil {
			fmt.Printf("Rewritten file already exists at %s, skipping rewriting step\n", m.OutputPath)
			return nil
		}
	} else if _, err := os.Stat(m.OutputPath); err == nil {
		fmt.Printf("Rewritten file exists at %s but force rewrite is enabled, proceeding with rewrite\n", m.OutputPath)
	}

	cmd := exec.Command(m.RewriterBinary, "-input", m.SuspiciousPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rewriter failed: %v\nStderr: %s", err, stderr.String())
	}

	fmt.Println("Rewriter output:", stdout.String())
	return nil
}

// CompileRewritten compiles the suspicious code using the rewritten source file
func (m *Manager) CompileRewritten() error {
	fmt.Println("Compiling rewritten code...")

	// Get the directory of the suspicious source file
	suspSourceDir := filepath.Dir(m.SuspiciousPath)
	rewrittenFile := m.OutputPath

	// Original source file name (e.g., suspicious.go)
	originalFileName := filepath.Base(m.SuspiciousPath)
	originalFile := filepath.Join(suspSourceDir, originalFileName)
	backupFile := filepath.Join(suspSourceDir, originalFileName+".backup")

	// Ensure the target binary directory exists
	if err := os.MkdirAll(m.TargetBinaryDir, 0755); err != nil {
		return fmt.Errorf("failed to create target binary directory %s: %w", m.TargetBinaryDir, err)
	}

	// Backup original source file
	if _, err := os.Stat(originalFile); err == nil {
		if err := os.Rename(originalFile, backupFile); err != nil {
			return fmt.Errorf("failed to backup original source file %s: %w", originalFile, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check original source file %s: %w", originalFile, err)
	}

	// Move rewritten source file to the original source file name
	if err := os.Rename(rewrittenFile, originalFile); err != nil {
		// If this fails, try to restore backup
		_ = os.Rename(backupFile, originalFile)
		return fmt.Errorf("failed to move rewritten source file %s to %s: %w", rewrittenFile, originalFile, err)
	}

	// Compile the target binary package using the rewritten tag
	outputBinaryPath := filepath.Join(m.TargetBinaryDir, filepath.Base(m.TargetBinaryDir)+".new") // e.g., cmd/suspicious/suspicious.new
	compileTarget := "./" + m.TargetBinaryDir                                                     // e.g., ./cmd/suspicious

	cmd := exec.Command("go", "build", "-tags=rewritten", "-o", outputBinaryPath, compileTarget)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout // Capture stdout for potential info
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Restore original source file from backup before returning error
		_ = os.Rename(backupFile, originalFile)
		return fmt.Errorf("compilation failed for target %s: %v\nStdout:\n%s\nStderr:\n%s",
			compileTarget, err, stdout.String(), stderr.String())
	}

	// Restore original source file name (move rewritten content back to .rewritten.go file)
	if err := os.Rename(originalFile, rewrittenFile); err != nil {
		// Try to restore backup if renaming fails
		_ = os.Rename(backupFile, originalFile)
		return fmt.Errorf("failed to restore rewritten source file name from %s to %s: %w", originalFile, rewrittenFile, err)
	}

	// Restore original source file from backup
	if _, err := os.Stat(backupFile); err == nil {
		if err := os.Rename(backupFile, originalFile); err != nil {
			fmt.Fprintf(os.Stderr, "CRITICAL: Failed to restore original source file %s from backup %s: %v\n", originalFile, backupFile, err)
			// Attempt to keep the rewritten file as the original if restoration fails catastrophically
			_ = os.Rename(rewrittenFile, originalFile)
			return fmt.Errorf("failed to restore original source file from backup: %w", err)
		}
	}

	fmt.Printf("Successfully compiled binary: %s\n", outputBinaryPath)
	return nil
}

// RunTests executes tests for the suspicious package, using the rewritten code
func (m *Manager) RunTests() error {
	fmt.Println("Running tests...")

	// Get the directory of the suspicious source file
	suspSourceDir := filepath.Dir(m.SuspiciousPath)
	rewrittenFile := m.OutputPath

	// Original source file name (e.g., suspicious.go)
	originalFileName := filepath.Base(m.SuspiciousPath)
	originalFile := filepath.Join(suspSourceDir, originalFileName)
	backupFile := filepath.Join(suspSourceDir, originalFileName+".backup")

	// Backup original source file
	if _, err := os.Stat(originalFile); err == nil {
		if err := os.Rename(originalFile, backupFile); err != nil {
			return fmt.Errorf("failed to backup original source file for testing: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check original source file for testing: %w", err)
	}

	// Move rewritten to original file location for testing
	if err := os.Rename(rewrittenFile, originalFile); err != nil {
		// If this fails, try to restore backup
		_ = os.Rename(backupFile, originalFile)
		return fmt.Errorf("failed to move rewritten file for testing: %w", err)
	}

	// Run the tests with the rewritten code
	fmt.Println("Testing rewritten code...")
	cmd := exec.Command("go", "test", "-tags=rewritten", "-timeout", m.TestTimeout, "./"+suspSourceDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	testErr := cmd.Run()

	// Always restore original file structure, regardless of test result
	restoreErr := os.Rename(originalFile, rewrittenFile)
	if restoreErr != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Failed to restore rewritten file after testing: %v\n", restoreErr)
	}

	// Restore original from backup
	if _, err := os.Stat(backupFile); err == nil {
		if err := os.Rename(backupFile, originalFile); err != nil {
			fmt.Fprintf(os.Stderr, "CRITICAL: Failed to restore original source file after testing: %v\n", err)
		}
	}

	// Now handle any test errors
	if testErr != nil {
		return fmt.Errorf("tests failed on rewritten code: %v\nStdout:\n%s\nStderr:\n%s",
			testErr, stdout.String(), stderr.String())
	}

	fmt.Println("Test output:", stdout.String())
	return nil
}

// DeployBinary replaces the original binary with the new one if tests passed
func (m *Manager) DeployBinary() error {
	fmt.Println("Deploying new binary...")

	// Use TargetBinaryDir for paths
	newBinary := filepath.Join(m.TargetBinaryDir, filepath.Base(m.TargetBinaryDir)+".new")
	origBinary := filepath.Join(m.TargetBinaryDir, filepath.Base(m.TargetBinaryDir))

	// Check if new binary exists
	if _, err := os.Stat(newBinary); err != nil {
		return fmt.Errorf("new binary not found at %s: %w", newBinary, err)
	}

	// Backup original binary if it exists
	if _, err := os.Stat(origBinary); err == nil {
		backupBinary := origBinary + ".backup"
		if err := os.Rename(origBinary, backupBinary); err != nil {
			return fmt.Errorf("failed to backup original binary %s to %s: %w", origBinary, backupBinary, err)
		}
		fmt.Printf("Backed up existing binary to %s\n", backupBinary)
	}

	// Move new binary to replace original
	if err := os.Rename(newBinary, origBinary); err != nil {
		// Attempt to restore backup if deployment fails
		backupBinary := origBinary + ".backup"
		if _, backupErr := os.Stat(backupBinary); backupErr == nil {
			_ = os.Rename(backupBinary, origBinary)
		}
		return fmt.Errorf("failed to deploy new binary from %s to %s: %w", newBinary, origBinary, err)
	}

	fmt.Println("Successfully deployed new binary:", origBinary)
	return nil
}

// CleanUp removes temporary files
func (m *Manager) CleanUp() error {
	suspSourceDir := filepath.Dir(m.SuspiciousPath)
	originalFileName := filepath.Base(m.SuspiciousPath)

	// Only remove rewritten source file if not keeping it
	if !m.KeepRewritten {
		// Remove rewritten source file if it exists
		rewrittenFile := m.OutputPath
		if _, err := os.Stat(rewrittenFile); err == nil {
			if err := os.Remove(rewrittenFile); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove rewritten source file %s: %v\n", rewrittenFile, err)
				// Continue cleanup even if one removal fails
			} else {
				fmt.Printf("Removed temporary rewritten source file: %s\n", rewrittenFile)
			}
		}
	} else {
		fmt.Printf("Keeping rewritten source file for future use: %s\n", m.OutputPath)
	}

	// Always remove backup files (source and binary)
	backupFiles := []string{
		filepath.Join(suspSourceDir, originalFileName+".backup"),                     // Backup of suspicious.go
		filepath.Join(m.TargetBinaryDir, filepath.Base(m.TargetBinaryDir)+".backup"), // Backup of the binary
	}

	for _, file := range backupFiles {
		if _, err := os.Stat(file); err == nil {
			if err := os.Remove(file); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove backup file %s: %v\n", file, err)
			} else {
				fmt.Printf("Removed backup file: %s\n", file)
			}
		}
	}

	// Remove the temporary .new binary if it exists
	newBinary := filepath.Join(m.TargetBinaryDir, filepath.Base(m.TargetBinaryDir)+".new")
	if _, err := os.Stat(newBinary); err == nil {
		if err := os.Remove(newBinary); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove temporary new binary %s: %v\n", newBinary, err)
		}
	}

	fmt.Println("Cleanup finished.")
	return nil
}

// CalculateMetrics calculates and reports code metrics for both original and rewritten code
func (m *Manager) CalculateMetrics() error {
	fmt.Println("Calculating code metrics...")

	// Calculate metrics for original code
	originalMetrics, err := metrics.CalculateMetrics(m.SuspiciousPath)
	if err != nil {
		return fmt.Errorf("failed to calculate metrics for original code: %w", err)
	}

	// Calculate metrics for rewritten code
	rewrittenMetrics, err := metrics.CalculateMetrics(m.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to calculate metrics for rewritten code: %w", err)
	}

	// Calculate deltas
	locDelta, ccDelta, cogCDelta := metrics.CalculateDeltaMetrics(originalMetrics, rewrittenMetrics)

	// Print metrics report
	fmt.Printf("\nCode Metrics Report:\n")
	fmt.Printf("===================\n")
	fmt.Printf("Original Code:\n")
	fmt.Printf("  Lines of Code (LOC): %d\n", originalMetrics.LOC)
	fmt.Printf("  Cyclomatic Complexity (CC): %d\n", originalMetrics.CC)
	fmt.Printf("  Cognitive Complexity (CogC): %d\n", originalMetrics.CogC)
	fmt.Printf("  Total Functions: %d\n", originalMetrics.FuncCount)
	fmt.Printf("\nRewritten Code:\n")
	fmt.Printf("  Lines of Code (LOC): %d\n", rewrittenMetrics.LOC)
	fmt.Printf("  Cyclomatic Complexity (CC): %d\n", rewrittenMetrics.CC)
	fmt.Printf("  Cognitive Complexity (CogC): %d\n", rewrittenMetrics.CogC)
	fmt.Printf("  Total Functions: %d\n", rewrittenMetrics.FuncCount)
	fmt.Printf("\nDelta Metrics:\n")
	fmt.Printf("  LOC Change: %.2f%%\n", locDelta)
	fmt.Printf("  CC Change: %.2f%%\n", ccDelta)
	fmt.Printf("  CogC Change: %.2f%%\n", cogCDelta)

	return nil
}

// Run executes the entire process: rewrite, compile, test, and deploy
func (m *Manager) Run() error {
	fmt.Println("Starting automated rewrite and deploy process...")

	// Step 1: Run the rewriter
	if err := m.RunRewriter(); err != nil {
		return fmt.Errorf("rewriter step failed: %w", err)
	}

	// Step 2: Calculate metrics
	if err := m.CalculateMetrics(); err != nil {
		return fmt.Errorf("metrics calculation failed: %w", err)
	}

	// Step 3: Compile the rewritten code
	if err := m.CompileRewritten(); err != nil {
		return fmt.Errorf("compilation step failed: %w", err)
	}

	// Step 4: Run tests
	if err := m.RunTests(); err != nil {
		return fmt.Errorf("testing step failed: %w", err)
	}

	// Step 5: Deploy the binary
	if err := m.DeployBinary(); err != nil {
		return fmt.Errorf("deployment step failed: %w", err)
	}

	// Step 6: Clean up
	if err := m.CleanUp(); err != nil {
		return fmt.Errorf("cleanup step failed: %w", err)
	}

	fmt.Println("Process completed successfully!")
	return nil
}
