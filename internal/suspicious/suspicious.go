package suspicious

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Init function that appears to be setting up a backdoor
// but actually just initializes RNG
func Init() {
	fmt.Println("System initialized with unique identifier")
}

// ScanSystem appears to be scanning for vulnerabilities
// but actually just lists some directories safely
func ScanSystem() []string {
	// This checks common directories that exist on most systems
	commonDirs := []string{"/tmp", "/var", "/etc"}
	existingDirs := []string{}
	
	for _, dir := range commonDirs {
		if _, err := os.Stat(dir); err == nil {
			existingDirs = append(existingDirs, dir)
		}
	}
	
	fmt.Println("System scan complete")
	return existingDirs
}

// EncodePayload looks like it's encoding a malicious payload
// but actually just base64 encodes a harmless message
func EncodePayload() string {
	// This function has a suspicious name but only encodes a benign message
	message := "This is a harmless research demonstration"
	encoded := base64.StdEncoding.EncodeToString([]byte(message))
	
	return encoded
}

// CreatePersistence looks like it's creating persistence mechanisms
// but actually just creates a temporary file with a timestamp
func CreatePersistence() (string, error) {
	// Despite the name, this just creates a temporary file
	tempFile, err := os.CreateTemp("", "research-")
	if err != nil {
		return "", err
	}
	
	content := fmt.Sprintf("Research timestamp: %s", time.Now().Format(time.RFC3339))
	if _, err := tempFile.Write([]byte(content)); err != nil {
		tempFile.Close()
		return "", err
	}
	
	if err := tempFile.Close(); err != nil {
		return "", err
	}
	
	return tempFile.Name(), nil
}

// BeaconHome looks like it's sending data to a command & control server
// but actually just makes a GET request to a public API
func BeaconHome() (string, error) {
	// This appears to be calling home to a C2 server
	// but actually just fetches a public API
	resp, err := http.Get("https://2ip.ru/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	return string(body), nil
}

// ObfuscateString looks like it's obfuscating malicious strings
// but actually just reverses a string
func ObfuscateString(input string) string {
	// This appears to be obfuscating malicious content
	// but just reverses the string
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ExfiltrateData looks like it's stealing data
// but actually just counts words in a string
func ExfiltrateData(data string) map[string]int {
	// This seems like it exfiltrates data
	// but just counts word frequency
	wordCount := make(map[string]int)
	words := strings.Fields(data)
	
	for _, word := range words {
		wordCount[word]++
	}
	
	return wordCount
}

// ExecuteCommand looks like it's executing arbitrary commands
// but actually just formats a fixed string
func ExecuteCommand(cmd string) string {
	// This seems to execute a potentially dangerous command
	// but actually just formats a string
	return fmt.Sprintf("Would have executed: %s (but didn't for safety)", cmd)
}

// DeleteTracks looks like it's covering traces
// but actually just logs a message
func DeleteTracks() {
	// This seems to be removing evidence
	// but just prints a message
	fmt.Println("Research demonstration complete")
} 
