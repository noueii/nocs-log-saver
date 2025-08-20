package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

type ParseTestRequest struct {
	Logs string `json:"logs"`
}

type ParseTestResult struct {
	LineNumber int         `json:"line_number"`
	Content    string      `json:"content"`
	Success    bool        `json:"success"`
	EventType  string      `json:"event_type,omitempty"`
	EventData  interface{} `json:"event_data,omitempty"`
	Error      string      `json:"error,omitempty"`
}

type ParseTestResponse struct {
	TotalLines  int               `json:"total_lines"`
	ParsedCount int               `json:"parsed_count"`
	FailedCount int               `json:"failed_count"`
	Results     []ParseTestResult `json:"results"`
}

// Test logs for each event type
var testLogs = map[string]string{
	// ========== CORE EVENTS (cs2-log library) ==========
	
	// Combat Events
	"kill": `L 08/19/2025 - 19:03:31: "alker007<8><[U:1:869707820]><CT>" [-1987 1958 0] killed "NxS Sebo<6><[U:1:387734521]><TERRORIST>" [-1946 1416 88] with "m4a1_silencer"`,
	"kill_assist": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" assisted killing "Player2<2><[U:1:654321]><TERRORIST>"`,
	"attack": `L 08/19/2025 - 19:03:31: "NxS Sebo<6><[U:1:387734521]><TERRORIST>" [-1943 1425 88] attacked "alker007<8><[U:1:869707820]><CT>" [-1987 1956 0] with "inferno" (damage "4") (damage_armor "0") (health "53") (armor "98") (hitgroup "generic")`,
	"suicide": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" [-100 200 0] committed suicide with "world"`,
	"killed_by_bomb": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><TERRORIST>" [-100 200 0] was killed by the bomb.`,
	
	// Round Events
	"round_start": `L 08/19/2025 - 19:03:56: World triggered "Round_Start"`,
	"round_end": `L 08/19/2025 - 19:03:56: World triggered "Round_End"`,
	"round_restart": `L 08/19/2025 - 19:03:56: World triggered "Restart_Round_(1_second)"`,
	"freeze_time_start": `L 08/19/2025 - 19:03:56: World triggered "Round_Freeze_End"`,
	"freeze_period_start": `L 08/19/2025 - 19:03:31: Starting Freeze period`,
	
	// Match Events
	"match_start": `L 08/19/2025 - 19:03:56: World triggered "Match_Start" on "de_dust2"`,
	"match_end": `L 08/19/2025 - 19:03:56: World triggered "Game_Over"`,  // Proper match_end trigger
	"game_commencing": `L 08/19/2025 - 19:03:56: World triggered "Game_Commencing"`,
	
	// Player Connection Events
	"player_connect": `L 08/19/2025 - 19:04:10: "SHESKY<7><[U:1:215888626]><>" connected, address "192.168.1.100:27005"`,
	"player_disconnect": `L 08/19/2025 - 19:04:10: "NxS Sebo<6><[U:1:387734521]><TERRORIST>" disconnected (reason "NETWORK_DISCONNECT_DISCONNECT_BY_USER")`,
	"player_entered": `L 08/19/2025 - 19:04:10: "SHESKY<7><[U:1:215888626]><>" entered the game`,
	
	// Communication Events (cs2-log PlayerSay)
	"chat": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" say "nice shot"`,
	
	// Team Events
	"team_switch": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]>" switched from team <Unassigned> to <TERRORIST>`,
	"team_scored": `L 08/19/2025 - 19:03:56: Team "CT" scored "16" with "5" players`,
	"team_notice": `L 08/19/2025 - 19:03:56: Team "TERRORIST" triggered "SFUI_Notice_Terrorists_Win"`,
	
	// Bomb Events
	"bomb_planted": `L 08/19/2025 - 19:03:31: "NxS Sebo<6><[U:1:387734521]><TERRORIST>" [100 200 0] planted the bomb at bombsite B`,
	"bomb_defused": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" defused the bomb`,
	"bomb_begin_defuse": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" triggered "Begin_Bomb_Defuse_With_Kit"`,
	"bomb_got": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><TERRORIST>" triggered "Got_The_Bomb"`,
	"bomb_dropped": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><TERRORIST>" dropped the bomb`,
	
	// Economy Events
	"purchase": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" purchased "m4a1"`,
	"money_change": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" money change 2700-1100 = $1600 (tracked) (purchase: item_assaultsuit)`,
	
	// Item Events
	"picked_up": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" picked up "ak47"`,
	"dropped": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" dropped "m4a1"`,
	
	// Grenade Events
	"grenade_thrown": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><TERRORIST>" threw smokegrenade [100 200 0] flashbang entindex 234)`,
	"blinded": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" blinded for 3.45 by "Player2<2><[U:1:654321]><TERRORIST>" from flashbang entindex 234`,
	"projectile_spawned": `L 08/19/2025 - 19:03:31: Molotov projectile spawned at 100.000000 200.000000 0.000000, velocity 500.000000 600.000000 100.000000`,
	
	// ========== EXTENDED EVENTS (Unknown classification) ==========
	
	// Buy Zone Events
	"left_buyzone": `L 08/19/2025 - 19:04:10: "NxS Sebo<6><[U:1:387734521]><TERRORIST>" left buyzone with [ weapon_knife weapon_usp_silencer kevlar(100) ]`,
	
	// Validation Events
	"userid_validated": `L 08/19/2025 - 19:04:10: "SHESKY<7><[U:1:215888626]><>" STEAM USERID validated`,
	
	// Achievement/Award Events
	"accolade_final_3k": `L 08/19/2025 - 19:03:31: ACCOLADE, FINAL: {3k}, NxS Sebo<6>, VALUE: 2.000000, POS: 1, SCORE: 53.333336`,
	"accolade_final_4k": `L 08/19/2025 - 19:03:31: ACCOLADE, FINAL: {4k}, Player1<1>, VALUE: 1.000000, POS: 1, SCORE: 70.000000`,
	"accolade_final_5k": `L 08/19/2025 - 19:03:31: ACCOLADE, FINAL: {5k}, Player1<1>, VALUE: 1.000000, POS: 1, SCORE: 100.000000`,
	"accolade_final_uniqueweaponkills": `L 08/19/2025 - 19:03:31: ACCOLADE, FINAL: {uniqueweaponkills}, SHESKY<7>, VALUE: 15.000000, POS: 1, SCORE: 70.000000`,
	"accolade_final_taserkills": `L 08/19/2025 - 19:03:31: ACCOLADE, FINAL: {taserkills}, alker007<8>, VALUE: 1.000000, POS: 1, SCORE: 26.666668`,
	"accolade_final_mvp": `L 08/19/2025 - 19:03:31: ACCOLADE, FINAL: {mvp}, Player1<1>, VALUE: 5.000000, POS: 1, SCORE: 80.000000`,
	
	// Match Status Events
	"match_status_score": `L 08/19/2025 - 19:03:56: MatchStatus: Score: 17:19 on map "de_dust2" RoundsPlayed: 36`,
	"match_status_teams": `L 08/19/2025 - 19:03:56: MatchStatus: Team playing "TERRORIST": team_SHESKY`,
	
	// Pause Events
	"match_pause_enabled": `L 08/19/2025 - 19:03:31: Match pause is enabled - mp_pause_match`,
	"match_pause_disabled": `L 08/19/2025 - 19:03:31: Match pause is disabled - TimeOutTs`,
	"match_unpause": `L 08/19/2025 - 19:03:31: Match unpaused`,
	
	// Debug Events
	"throw_debug_molotov": `L 08/19/2025 - 19:03:31: "NxS Sebo<6><[U:1:387734521]><TERRORIST>" sv_throw_molotov -1943.109 1620.291 94.267 0.000 0.000 0.000`,
	"throw_debug_smoke": `L 08/19/2025 - 19:03:31: "SHESKY<7><[U:1:215888626]><TERRORIST>" sv_throw_smokegrenade -465.938 1875.818 -68.629 0.000 0.000 0.000`,
	"throw_debug_flash": `L 08/19/2025 - 19:03:31: "SHESKY<7><[U:1:215888626]><TERRORIST>" sv_throw_flashgrenade -271.017 1354.833 -46.747 0.000 0.000 0.000`,
	"throw_debug_he": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" sv_throw_hegrenade 100.000 200.000 0.000 0.000 0.000 0.000`,
	
	// Server Configuration Events
	"server_cvar": `L 08/19/2025 - 19:03:31: server_cvar: "mp_freezetime" "20"`,
	"cvar_maxrounds": `L 08/19/2025 - 19:03:31: "mp_maxrounds" = "24"`,
	"cvar_overtime": `L 08/19/2025 - 19:03:31: "mp_overtime_enable" = "1"`,
	"cvar_freezetime": `L 08/19/2025 - 19:03:31: "mp_freezetime" = "15"`,
	"cvar_tournament": `L 08/19/2025 - 19:03:31: "mp_tournament" = "1"`,
	"cvar_mp_setting": `L 08/19/2025 - 19:03:31: "mp_winlimit" = "0"`,
	
	// Chat Events (Unknown type)
	"chat_pause_command": `L 08/19/2025 - 19:03:31: "brotacel<4><[U:1:210708726]><Spectator>" say ".pause"`,
	"chat_restore_command": `L 08/19/2025 - 19:03:31: "brotacel<4><[U:1:210708726]><Spectator>" say ".restore 35"`,
	"chat_ready_command": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" say ".ready"`,
	"chat_gg": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" say "gg wp"`,
	"chat_message": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" say "nice round guys"`,
	
	// Game State Events
	"game_over_competitive": `L 08/19/2025 - 19:03:31: Game Over: competitive de_dust2 score 17:19 after 49 min`,
	"game_over_casual": `L 08/19/2025 - 19:03:31: Game Over: casual de_mirage score 8:5 after 20 min`,
	// Note: freeze_period_start is handled by cs2log.FreezTimeStart, not Unknown type
	
	// Map Events
	"loading_map": `L 08/19/2025 - 19:03:31: Loading map "de_mirage"`,
	"started_map": `L 08/19/2025 - 19:03:31: Started map "de_dust2"`,
	
	// Team Events (Extended)
	"team_playing": `L 08/19/2025 - 19:03:31: Team playing "TERRORIST": team_xHaPPy_`,
	
	// Log File Events
	"log_file_started": `L 08/19/2025 - 19:03:31: Log file started (file "logs/2025_08_19_181143.log") (game "csgo") (version "10521")`,
	"log_file_closed": `L 08/19/2025 - 19:03:31: Log file closed`,
	
	// RCON Events
	"rcon_command": `L 08/19/2025 - 19:03:31: rcon from "192.168.1.100:12345": command "mp_pause_match 1"`,
	
	// Triggered Events
	"trigger_warmup-start": `L 08/19/2025 - 19:03:31: World triggered "Warmup_Start"`,
	"trigger_match-reloaded": `L 08/19/2025 - 19:03:31: World triggered "Match_Reloaded" on "de_dust2"`,
	"trigger_sfui-notice-round-draw": `L 08/19/2025 - 19:03:31: World triggered "SFUI_Notice_Round_Draw" (CT "6") (T "5")`,
	
	// Bomb Events (Extended)
	"bomb_begin_plant": `L 08/19/2025 - 19:03:31: "NxS Sebo<6><[U:1:387734521]><TERRORIST>" triggered "Bomb_Begin_Plant" at bombsite B`,
	"bomb_planted_trigger": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><TERRORIST>" triggered "Bomb_Planted"`,
	"bomb_defused_trigger": `L 08/19/2025 - 19:03:31: "Player1<1><[U:1:123456]><CT>" triggered "Bomb_Defused"`,
}

func main() {
	fmt.Println("=== CS2 Event Types Comprehensive Test ===")
	fmt.Printf("Testing %d event types...\n\n", len(testLogs))
	
	// Prepare logs for testing
	var allLogs []string
	eventOrder := make([]string, 0, len(testLogs))
	
	// Sort event types for consistent output
	for eventType := range testLogs {
		eventOrder = append(eventOrder, eventType)
	}
	sort.Strings(eventOrder)
	
	// Build the log string
	for _, eventType := range eventOrder {
		allLogs = append(allLogs, testLogs[eventType])
	}
	
	// Create request
	req := ParseTestRequest{
		Logs: strings.Join(allLogs, "\n"),
	}
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		return
	}
	
	// Send request to parse-test endpoint
	resp, err := http.Post("http://localhost:9090/api/parse-test", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}
	
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error response (status %d): %s\n", resp.StatusCode, string(body))
		return
	}
	
	// Parse response
	var parseResp ParseTestResponse
	if err := json.Unmarshal(body, &parseResp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}
	
	// Analyze results
	fmt.Printf("=== Test Results ===\n")
	fmt.Printf("Total Lines: %d\n", parseResp.TotalLines)
	fmt.Printf("Parsed: %d\n", parseResp.ParsedCount)
	fmt.Printf("Failed: %d\n\n", parseResp.FailedCount)
	
	// Check each expected event type
	resultsByType := make(map[string]bool)
	actualTypes := make(map[string]string) // maps expected to actual
	
	for i, result := range parseResp.Results {
		if i < len(eventOrder) {
			expectedType := eventOrder[i]
			actualType := result.EventType
			
			if result.Success {
				resultsByType[expectedType] = actualType == expectedType
				actualTypes[expectedType] = actualType
			} else {
				resultsByType[expectedType] = false
				actualTypes[expectedType] = "PARSE_FAILED: " + result.Error
			}
		}
	}
	
	// Display results by category
	categories := map[string][]string{
		"Combat Events": {"kill", "kill_assist", "attack", "suicide", "killed_by_bomb"},
		"Round Events": {"round_start", "round_end", "round_restart", "freeze_time_start", "freeze_period_start"},
		"Match Events": {"match_start", "match_end", "game_commencing", "game_over_competitive", "game_over_casual"},
		"Player Connection": {"player_connect", "player_disconnect", "player_entered", "userid_validated"},
		"Communication": {"chat", "chat_message", "chat_pause_command", "chat_restore_command", "chat_ready_command", "chat_gg"},
		"Team Events": {"team_switch", "team_scored", "team_notice", "team_playing"},
		"Bomb Events": {"bomb_planted", "bomb_defused", "bomb_begin_defuse", "bomb_got", "bomb_dropped", "bomb_begin_plant", "bomb_planted_trigger", "bomb_defused_trigger"},
		"Economy": {"purchase", "money_change", "left_buyzone"},
		"Items": {"picked_up", "dropped"},
		"Grenades": {"grenade_thrown", "blinded", "projectile_spawned", "throw_debug_molotov", "throw_debug_smoke", "throw_debug_flash", "throw_debug_he"},
		"Match Status": {"match_status_score", "match_status_teams"},
		"Pause": {"match_pause_enabled", "match_pause_disabled", "match_unpause"},
		"Server Config": {"server_cvar", "cvar_maxrounds", "cvar_overtime", "cvar_freezetime", "cvar_tournament", "cvar_mp_setting"},
		"Achievements": {"accolade_final_3k", "accolade_final_4k", "accolade_final_5k", "accolade_final_uniqueweaponkills", "accolade_final_taserkills", "accolade_final_mvp"},
		"Map Events": {"loading_map", "started_map"},
		"Logging": {"log_file_started", "log_file_closed"},
		"Admin": {"rcon_command"},
		"Triggered": {"trigger_warmup-start", "trigger_match-reloaded", "trigger_sfui-notice-round-draw"},
	}
	
	passCount := 0
	failCount := 0
	
	for category, events := range categories {
		fmt.Printf("\n=== %s ===\n", category)
		for _, eventType := range events {
			if success, exists := resultsByType[eventType]; exists {
				status := "❌ FAIL"
				if success {
					status = "✅ PASS"
					passCount++
				} else {
					failCount++
				}
				
				actual := actualTypes[eventType]
				if eventType == actual {
					fmt.Printf("%s %-40s -> %s\n", status, eventType, eventType)
				} else {
					fmt.Printf("%s %-40s -> %s (expected: %s)\n", status, eventType, actual, eventType)
				}
			} else {
				fmt.Printf("⚠️  %-40s -> NOT TESTED\n", eventType)
			}
		}
	}
	
	// Summary
	fmt.Printf("\n=== SUMMARY ===\n")
	fmt.Printf("Total Event Types Tested: %d\n", len(testLogs))
	fmt.Printf("✅ Passed: %d\n", passCount)
	fmt.Printf("❌ Failed: %d\n", failCount)
	successRate := float64(passCount) / float64(len(testLogs)) * 100
	fmt.Printf("Success Rate: %.1f%%\n", successRate)
	
	// List any unrecognized events
	fmt.Printf("\n=== Unrecognized Event Types ===\n")
	hasUnrecognized := false
	for expected, actual := range actualTypes {
		if strings.Contains(actual, "unknown") || strings.Contains(actual, "unrecognized") {
			fmt.Printf("%-40s -> %s\n", expected, actual)
			hasUnrecognized = true
		}
	}
	if !hasUnrecognized {
		fmt.Println("None! All events were properly categorized.")
	}
}