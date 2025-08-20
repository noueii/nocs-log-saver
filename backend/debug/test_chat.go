package main

import (
	"fmt"
	cs2log "github.com/janstuemmel/cs2-log"
)

func main() {
	// Test chat events
	chats := []string{
		`L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" say "nice shot"`,
		`L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" say ".ready"`,
		`L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" say "gg wp"`,
		`L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" say ".pause"`,
	}

	for _, chat := range chats {
		parsed, err := cs2log.Parse(chat)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Type: %T\n", parsed)
		
		if playerSay, ok := parsed.(cs2log.PlayerSay); ok {
			fmt.Printf("  Player: %s\n", playerSay.Player.Name)
			fmt.Printf("  Text: %s\n", playerSay.Text)
		}
	}

	// Test freeze events
	fmt.Println("\n=== Freeze Events ===")
	freezeEvents := []string{
		`L 08/19/2025 - 19:03:56: World triggered "Round_Freeze_End"`,
		`L 08/19/2025 - 19:03:31: Starting Freeze period`,
	}

	for _, event := range freezeEvents {
		parsed, err := cs2log.Parse(event)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("%s -> Type: %T\n", event, parsed)
	}
}