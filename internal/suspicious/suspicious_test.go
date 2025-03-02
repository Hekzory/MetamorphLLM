package suspicious

import (
	"os"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	// Just ensure it doesn't panic
	Init()
}

func TestScanSystem(t *testing.T) {
	dirs := ScanSystem()
	
	// Ensure we get some directories
	if len(dirs) == 0 {
		t.Error("Expected to find at least one existing directory")
	}
	
	// Ensure the directories exist
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("Directory %s should exist but was not accessible", dir)
		}
	}
}

func TestEncodePayload(t *testing.T) {
	encoded := EncodePayload()
	
	// Ensure we got a non-empty string
	if len(encoded) == 0 {
		t.Error("Expected a non-empty encoded payload")
	}
	
	// Ensure the string is base64-encoded (it should only contain valid base64 characters)
	validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	for _, c := range encoded {
		if !strings.ContainsRune(validChars, c) {
			t.Errorf("Encoded payload contains invalid base64 character: %c", c)
			break
		}
	}
}

func TestCreatePersistence(t *testing.T) {
	filename, err := CreatePersistence()
	
	// Ensure no error occurred
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	// Ensure the file was created
	if _, err := os.Stat(filename); err != nil {
		t.Errorf("File %s should exist but was not accessible", filename)
	}
	
	// Clean up
	_ = os.Remove(filename)
}

func TestObfuscateString(t *testing.T) {
	original := "hello world"
	obfuscated := ObfuscateString(original)
	expected := "dlrow olleh"
	
	if obfuscated != expected {
		t.Errorf("Expected '%s', got '%s'", expected, obfuscated)
	}
}

func TestExfiltrateData(t *testing.T) {
	data := "one two two three three three"
	wordCount := ExfiltrateData(data)
	
	// Check counts
	if wordCount["one"] != 1 {
		t.Errorf("Expected count of 'one' to be 1, got %d", wordCount["one"])
	}
	if wordCount["two"] != 2 {
		t.Errorf("Expected count of 'two' to be 2, got %d", wordCount["two"])
	}
	if wordCount["three"] != 3 {
		t.Errorf("Expected count of 'three' to be 3, got %d", wordCount["three"])
	}
}

func TestExecuteCommand(t *testing.T) {
	cmd := "rm -rf /"
	result := ExecuteCommand(cmd)
	
	// Ensure the result contains the command but wasn't actually executed
	if !strings.Contains(result, cmd) {
		t.Errorf("Expected result to contain '%s', got '%s'", cmd, result)
	}
	if !strings.Contains(result, "didn't") {
		t.Errorf("Expected result to indicate command wasn't executed, got '%s'", result)
	}
}

func TestBeaconHome(t *testing.T) {
	response, err := BeaconHome()
	
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if !strings.Contains(response, "origin") {
		t.Errorf("Expected response to contain 'origin', got: %s", response)
	}
}
