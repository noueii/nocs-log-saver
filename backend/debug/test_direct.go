package main

import (
	"fmt"
	"strings"

	cs2log "github.com/janstuemmel/cs2-log"
)

func main() {
	// Test a few sample logs
	testLogs := []string{
		`08/19/2025 - 19:04:10.830 - "NxS Sebo<6><[U:1:387734521]><TERRORIST>" disconnected (reason "NETWORK_DISCONNECT_DISCONNECT_BY_USER")`,
		`08/19/2025 - 19:03:31.480 - "alker007<8><[U:1:869707820]><CT>" [-1987 1958 0] killed "NxS Sebo<6><[U:1:387734521]><TERRORIST>" [-1946 1416 88] with "m4a1_silencer"`,
		`08/19/2025 - 19:03:56.030 - World triggered "Round_End"`,
		`08/19/2025 - 19:02:49.214 - "SHESKY<7><[U:1:215888626]><TERRORIST>" picked up "smokegrenade"`,
		`08/19/2025 - 19:03:31.339 - "NxS Sebo<6><[U:1:387734521]><TERRORIST>" [-1943 1425 88] attacked "alker007<8><[U:1:869707820]><CT>" [-1987 1956 0] with "inferno" (damage "4") (damage_armor "0") (health "53") (armor "98") (hitgroup "generic")`,
	}

	for i, log := range testLogs {
		fmt.Printf("\n=== Test %d ===\n", i+1)
		fmt.Printf("Original: %s\n", log)
		
		// Add L prefix
		testLog := "L " + log
		fmt.Printf("With L prefix: %s\n", testLog)
		
		// Parse with cs2-log
		parsed, err := cs2log.Parse(testLog)
		if err != nil {
			fmt.Printf("❌ Parse error: %v\n", err)
			continue
		}
		
		fmt.Printf("✅ Parsed successfully\n")
		fmt.Printf("Type: %T\n", parsed)
		
		// Check what the actual type is
		switch msg := parsed.(type) {
		case *cs2log.PlayerDisconnected:
			fmt.Printf("Correctly identified as PlayerDisconnected\n")
		case *cs2log.PlayerKill:
			fmt.Printf("Correctly identified as PlayerKill\n")
		case *cs2log.WorldRoundEnd:
			fmt.Printf("Correctly identified as WorldRoundEnd\n")
		case *cs2log.PlayerPickedUp:
			fmt.Printf("Correctly identified as PlayerPickedUp\n")
		case *cs2log.PlayerAttack:
			fmt.Printf("Correctly identified as PlayerAttack\n")
		case *cs2log.Unknown:
			fmt.Printf("Parsed as Unknown - Raw: %s\n", msg.Raw)
		default:
			fmt.Printf("Unexpected type: %T\n", msg)
		}
	}
	
	// Test with milliseconds removal
	fmt.Println("\n=== Testing millisecond handling ===")
	logWithMs := "L 08/19/2025 - 19:03:31.480 - Test log"
	logWithoutMs := "L 08/19/2025 - 19:03:31: Test log"
	
	fmt.Printf("With ms: %s\n", logWithMs)
	fmt.Printf("Without ms: %s\n", logWithoutMs)
	
	// Check if time format is the issue
	parts := strings.SplitN(logWithMs, " - ", 3)
	if len(parts) >= 3 {
		fmt.Printf("Date: %s\n", parts[0])
		fmt.Printf("Time: %s\n", parts[1])
		fmt.Printf("Content: %s\n", parts[2])
	}
}