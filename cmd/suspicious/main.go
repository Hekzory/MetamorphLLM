package main

import (
	"fmt"
	"github.com/Hekzory/polymorphengine/internal/suspicious"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== Suspicious-Looking Program ===")
	fmt.Println("NOTE: This program is for research purposes and is completely harmless.")
	fmt.Println("      It demonstrates code that looks suspicious but does nothing malicious.")
	fmt.Println("=====================================")
	
	// Initialize the system
	suspicious.Init()
	
	// Perform operations that sound suspicious but are harmless
	fmt.Println("\n[*] Scanning system...")
	dirs := suspicious.ScanSystem()
	fmt.Printf("[+] Found %d target directories\n", len(dirs))
	
	fmt.Println("\n[*] Encoding payload...")
	encoded := suspicious.EncodePayload()
	fmt.Printf("[+] Payload prepared: %s\n", encoded)
	
	fmt.Println("\n[*] Creating persistence...")
	filename, err := suspicious.CreatePersistence()
	if err != nil {
		fmt.Printf("[-] Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[+] Established persistence at: %s\n", filename)
	defer os.Remove(filename) // Clean up
	
	fmt.Println("\n[*] Obfuscating communications...")
	message := "Communication channel established"
	obfuscated := suspicious.ObfuscateString(message)
	fmt.Printf("[+] Obfuscated message: %s\n", obfuscated)
	
	fmt.Println("\n[*] Exfiltrating data...")
	sampleData := "This is sample data that would be exfiltrated in a real attack"
	wordCounts := suspicious.ExfiltrateData(sampleData)
	fmt.Println("[+] Data analysis complete. Word frequencies:")
	for word, count := range wordCounts {
		if count > 1 {
			fmt.Printf("    - %s: %d occurrences\n", word, count)
		}
	}
	
	fmt.Println("\n[*] Executing command...")
	cmd := "echo 'System compromised'"
	result := suspicious.ExecuteCommand(cmd)
	fmt.Printf("[+] Result: %s\n", result)
	
	fmt.Println("\n[*] Beaconing home...")
	response, err := suspicious.BeaconHome()
	if err != nil {
		fmt.Printf("[-] Error: %v\n", err)
	} else {
		fmt.Printf("[+] Response received: %s\n", strings.TrimSpace(response))
	}
	
	fmt.Println("\n[*] Cleaning up tracks...")
	suspicious.DeleteTracks()
	
	fmt.Println("\n=== Program Finished ===")
	fmt.Printf("Executed at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Println("Remember: This was just a demonstration. No actual malicious activity occurred.")
} 