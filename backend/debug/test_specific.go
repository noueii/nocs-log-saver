package main

import (
	"fmt"
	cs2log "github.com/janstuemmel/cs2-log"
)

func main() {
	// Test specific problematic logs
	tests := []struct {
		name     string
		log      string
		expected string
	}{
		{
			"bomb_planted",
			`L 08/19/2025 - 19:03:31: "NxS Sebo<6><[U:1:387734521]><TERRORIST>" [100 200 0] planted the bomb at bombsite B`,
			"bomb_planted",
		},
		{
			"bomb_defused", 
			`L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" defused the bomb`,
			"bomb_defused",
		},
		{
			"bomb_dropped",
			`L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><TERRORIST>" dropped the bomb`,
			"bomb_dropped",
		},
		{
			"rcon",
			`L 08/19/2025 - 19:03:31: rcon from "192.168.1.100:12345": command "mp_pause_match 1"`,
			"rcon_command",
		},
		{
			"freeze_time",
			`L 08/19/2025 - 19:03:56: World triggered "Round_Freeze_End"`,
			"freeze_time_start",
		},
		{
			"team_notice",
			`L 08/19/2025 - 19:03:56: Team "TERRORIST" triggered "SFUI_Notice_Terrorists_Win"`,
			"team_notice",
		},
	}

	for _, test := range tests {
		fmt.Printf("\n=== Testing %s ===\n", test.name)
		fmt.Printf("Log: %s\n", test.log)
		
		parsed, err := cs2log.Parse(test.log)
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			continue
		}
		
		fmt.Printf("Type: %T\n", parsed)
		
		// Check specific types
		switch v := parsed.(type) {
		case cs2log.PlayerBombPlanted:
			fmt.Println("✅ Recognized as PlayerBombPlanted")
		case cs2log.PlayerBombDefused:
			fmt.Println("✅ Recognized as PlayerBombDefused")
		case cs2log.PlayerBombDropped:
			fmt.Println("✅ Recognized as PlayerBombDropped")
		case cs2log.Unknown:
			fmt.Printf("⚠️ Parsed as Unknown, raw: %s\n", v.Raw)
		default:
			fmt.Printf("❌ Unexpected type: %T\n", v)
		}
	}
}