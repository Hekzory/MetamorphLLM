package metrics

import (
	"os"
	"testing"
)

func TestCalculateMetrics(t *testing.T) {
	// Create a temporary test file
	testCode := `package test

func simple() {
	fmt.Println("Hello")
}

func complex() {
	if true {
		for i := 0; i < 10; i++ {
			if i%2 == 0 {
				fmt.Println(i)
			}
		}
	}
}

func nested() {
	if true {
		if false {
			fmt.Println("Never")
		}
	}
}`

	tmpFile, err := os.CreateTemp("", "test_*.go")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(testCode)); err != nil {
		t.Fatalf("Failed to write test code: %v", err)
	}
	tmpFile.Close()

	// Calculate metrics
	metrics, err := CalculateMetrics(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	// Verify metrics
	if metrics.FuncCount != 3 {
		t.Errorf("Expected 3 functions, got %d", metrics.FuncCount)
	}

	if metrics.LOC < 10 {
		t.Errorf("Expected more than 10 lines of code, got %d", metrics.LOC)
	}

	if metrics.CC < 5 {
		t.Errorf("Expected cyclomatic complexity > 5, got %d", metrics.CC)
	}

	if metrics.CogC < 5 {
		t.Errorf("Expected cognitive complexity > 5, got %d", metrics.CogC)
	}
}

func TestCalculateFunctionalEquivalence(t *testing.T) {
	tests := []struct {
		passedTests int
		totalTests  int
		expected    float64
	}{
		{5, 10, 50.0},
		{0, 10, 0.0},
		{10, 10, 100.0},
		{0, 0, 0.0},
	}

	for _, test := range tests {
		result := CalculateFunctionalEquivalence(test.passedTests, test.totalTests)
		if result != test.expected {
			t.Errorf("CalculateFunctionalEquivalence(%d, %d) = %.2f, want %.2f",
				test.passedTests, test.totalTests, result, test.expected)
		}
	}
}

func TestCalculateDeltaMetrics(t *testing.T) {
	original := &Metrics{
		LOC:  100,
		CC:   10,
		CogC: 15,
	}

	metamorphic := &Metrics{
		LOC:  120,
		CC:   12,
		CogC: 18,
	}

	locDelta, ccDelta, cogCDelta := CalculateDeltaMetrics(original, metamorphic)

	expectedLocDelta := 20.0  // (120-100)/100 * 100
	expectedCCDelta := 20.0   // (12-10)/10 * 100
	expectedCogCDelta := 20.0 // (18-15)/15 * 100

	if locDelta != expectedLocDelta {
		t.Errorf("LOC delta = %.2f%%, want %.2f%%", locDelta, expectedLocDelta)
	}

	if ccDelta != expectedCCDelta {
		t.Errorf("CC delta = %.2f%%, want %.2f%%", ccDelta, expectedCCDelta)
	}

	if cogCDelta != expectedCogCDelta {
		t.Errorf("CogC delta = %.2f%%, want %.2f%%", cogCDelta, expectedCogCDelta)
	}
}
