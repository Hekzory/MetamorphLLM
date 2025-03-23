package main

import (
	"fmt"
	"github.com/Hekzory/MetamorphLLM/internal/suspicious"
	"os"
	"strings"
	"time"
)

func main() {
	printHeader()
	
	// Initialize the system
	suspicious.Init()
	
	// Run all suspicious-looking but harmless operations
	runScanSystem()
	runEncodePayload()
	filename := runCreatePersistence()
	defer os.Remove(filename) // Clean up
	
	runObfuscate()
	runExfiltrate()
	runExecuteCommand()
	runBeacon()
	
	// Clean up
	fmt.Println("\n[*] Cleaning up tracks...")
	suspicious.DeleteTracks()
	
	printFooter()
}

func printHeader() {
	fmt.Println("=== Suspicious-Looking Program ===")
	fmt.Println("NOTE: This program is for research purposes and is completely harmless.")
	fmt.Println("      It demonstrates code that looks suspicious but does nothing malicious.")
	fmt.Println("=====================================")
}

func printFooter() {
	fmt.Println("\n=== Program Finished ===")
	fmt.Printf("Executed at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Println("Remember: This was just a demonstration. No actual malicious activity occurred.")
}

func runScanSystem() {
	fmt.Println("\n[*] Scanning system...")
	dirs := suspicious.ScanSystem()
	fmt.Printf("[+] Found %d target directories\n", len(dirs))
}

func runEncodePayload() {
	fmt.Println("\n[*] Encoding payload...")
	encoded := suspicious.EncodePayload()
	fmt.Printf("[+] Payload prepared: %s\n", encoded)
}

func runCreatePersistence() string {
	fmt.Println("\n[*] Creating persistence...")
	filename, err := suspicious.CreatePersistence()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[-] Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[+] Established persistence at: %s\n", filename)
	return filename
}

func runObfuscate() {
	fmt.Println("\n[*] Obfuscating communications...")
	message := "Communication channel established"
	obfuscated := suspicious.ObfuscateString(message)
	fmt.Printf("[+] Obfuscated message: %s\n", obfuscated)
}

func runExfiltrate() {
	fmt.Println("\n[*] Exfiltrating data...")
	sampleData := "This is sample data that would be exfiltrated in a real attack"
	wordCounts := suspicious.ExfiltrateData(sampleData)
	fmt.Println("[+] Data analysis complete. Word frequencies:")
	
	// Print words that occur more than once
	for word, count := range wordCounts {
		if count > 1 {
			fmt.Printf("    - %s: %d occurrences\n", word, count)
		}
	}
}

func runExecuteCommand() {
	fmt.Println("\n[*] Executing command...")
	cmd := "echo 'System compromised'"
	result := suspicious.ExecuteCommand(cmd)
	fmt.Printf("[+] Result: %s\n", result)
}

func runBeacon() {
	fmt.Println("\n[*] Beaconing home...")
	response, err := suspicious.BeaconHome()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[-] Error: %v\n", err)
		return
	}
	fmt.Printf("[+] Response received: %s\n", strings.TrimSpace(response))
} 