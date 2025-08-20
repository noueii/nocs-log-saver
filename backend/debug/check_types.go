package main

import (
	"fmt"
	"reflect"
	"strings"

	cs2log "github.com/janstuemmel/cs2-log"
)

func main() {
	// Test log that should parse as PlayerDisconnected
	testLog := `L 08/19/2025 - 19:04:10: "NxS Sebo<6><[U:1:387734521]><TERRORIST>" disconnected (reason "NETWORK_DISCONNECT_DISCONNECT_BY_USER")`
	
	parsed, err := cs2log.Parse(testLog)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}
	
	fmt.Printf("Parsed successfully!\n")
	fmt.Printf("Type: %T\n", parsed)
	fmt.Printf("Type Name: %s\n", reflect.TypeOf(parsed).String())
	fmt.Printf("Kind: %s\n", reflect.TypeOf(parsed).Kind())
	
	// Check if it's a pointer
	if reflect.TypeOf(parsed).Kind() == reflect.Ptr {
		fmt.Printf("Element Type: %s\n", reflect.TypeOf(parsed).Elem().String())
	}
	
	// Try type assertion
	switch v := parsed.(type) {
	case *cs2log.PlayerDisconnected:
		fmt.Println("✅ Correctly identified as *cs2log.PlayerDisconnected")
	case cs2log.PlayerDisconnected:
		fmt.Println("✅ Identified as cs2log.PlayerDisconnected (no pointer)")
	default:
		fmt.Printf("❌ Not recognized, actual type: %T\n", v)
		
		// Check package path
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		fmt.Printf("Package path: %s\n", t.PkgPath())
		fmt.Printf("Type name: %s\n", t.Name())
		
		// Check if it's actually the same type but different import
		typeName := fmt.Sprintf("%T", v)
		if strings.Contains(typeName, "PlayerDisconnected") {
			fmt.Println("⚠️  Type contains 'PlayerDisconnected' but type assertion failed!")
			fmt.Println("This might be an import path issue.")
		}
	}
	
	// List all available types in cs2log package
	fmt.Println("\n=== Testing Other Event Types ===")
	
	testLogs := map[string]string{
		"PlayerKill": `L 08/19/2025 - 19:03:31: "alker007<8><[U:1:869707820]><CT>" [-1987 1958 0] killed "NxS Sebo<6><[U:1:387734521]><TERRORIST>" [-1946 1416 88] with "m4a1_silencer"`,
		"WorldRoundEnd": `L 08/19/2025 - 19:03:56: World triggered "Round_End"`,
		"PlayerAttack": `L 08/19/2025 - 19:03:31: "NxS Sebo<6><[U:1:387734521]><TERRORIST>" [-1943 1425 88] attacked "alker007<8><[U:1:869707820]><CT>" [-1987 1956 0] with "inferno" (damage "4") (damage_armor "0") (health "53") (armor "98") (hitgroup "generic")`,
	}
	
	for expectedType, logLine := range testLogs {
		parsed, err := cs2log.Parse(logLine)
		if err != nil {
			fmt.Printf("%s: Parse error: %v\n", expectedType, err)
			continue
		}
		fmt.Printf("%s: Type is %T\n", expectedType, parsed)
	}
}