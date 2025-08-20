# CS2 Event Types Documentation

This document provides a comprehensive reference for all CS2 event types recognized by the log parser. Events are organized by category with descriptions, use cases, and example log entries.

## Table of Contents
- [Core Events (Parsed by cs2-log library)](#core-events-parsed-by-cs2-log-library)
- [Extended Events (Custom Classification)](#extended-events-custom-classification)
  - [Buy Zone Events](#buy-zone-events)
  - [Validation Events](#validation-events)
  - [Achievement/Award Events](#achievementaward-events)
  - [Match Status Events](#match-status-events)
  - [Pause Events](#pause-events)
  - [Debug Events](#debug-events)
  - [Server Configuration Events](#server-configuration-events)
  - [Chat Events](#chat-events)
  - [Game State Events](#game-state-events)
  - [Map Events](#map-events)
  - [Triggered Events](#triggered-events)

---

## Core Events (Parsed by cs2-log library)

These events are directly parsed by the [cs2-log](https://github.com/janstuemmel/cs2-log) library and contain structured data.

### Combat Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `kill` | Player killed another player | `"Player1<1><STEAM_ID><CT>" killed "Player2<2><STEAM_ID><T>" with "ak47"` |
| `kill_assist` | Player assisted in a kill | `"Player1<1><STEAM_ID><CT>" assisted killing "Player2<2><STEAM_ID><T>"` |
| `attack` | Player damaged another player | `"Player1<1><STEAM_ID><T>" attacked "Player2<2><STEAM_ID><CT>" with "glock" (damage "27")` |
| `killed_by_bomb` | Player killed by bomb explosion | `"Player1<1><STEAM_ID><T>" was killed by the bomb` |
| `suicide` | Player committed suicide | `"Player1<1><STEAM_ID><CT>" committed suicide` |

### Round Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `round_start` | Round has started | `World triggered "Round_Start"` |
| `round_end` | Round has ended | `World triggered "Round_End"` |
| `round_restart` | Round is restarting | `World triggered "Round_Restart"` |
| `freeze_time_start` | Freeze time period started | `World triggered "Round_Freeze_End"` |

### Match Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `match_start` | Match has started | `World triggered "Match_Start" on "de_dust2"` |
| `match_end` / `trigger_game-over` | Match has ended | `World triggered "Game_Over"` |
| `game_commencing` | Game is about to start | `World triggered "Game_Commencing"` |

### Player Connection Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `player_connect` | Player connected to server | `"Player1<1><STEAM_ID><>" connected` |
| `player_disconnect` | Player disconnected from server | `"Player1<1><STEAM_ID><CT>" disconnected (reason "Disconnect")` |
| `player_entered` | Player entered the game | `"Player1<1><STEAM_ID><>" entered the game` |

### Communication Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `chat_message` | General player chat message | `"Player1<1><STEAM_ID><CT>" say "nice shot"` |
| `chat_pause_command` | Player requested pause | `"Player1<1><STEAM_ID><CT>" say ".pause"` |
| `chat_restore_command` | Player requested restore | `"Player1<1><STEAM_ID><CT>" say ".restore 35"` |
| `chat_ready_command` | Player signaled ready | `"Player1<1><STEAM_ID><CT>" say ".ready"` |
| `chat_gg` | Good game message | `"Player1<1><STEAM_ID><CT>" say "gg wp"` |

### Team Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `team_switch` | Player switched teams | `"Player1<1><STEAM_ID>" switched from team <CT> to <TERRORIST>` |
| `team_scored` | Team scored points | `Team "CT" scored "16"` |
| `team_notice` | Team-related notification | `Team "TERRORIST" triggered "SFUI_Notice_Terrorists_Win"` |

### Bomb Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `bomb_planted` | Bomb has been planted | `"Player1<1><STEAM_ID><T>" planted the bomb` |
| `bomb_defused` | Bomb has been defused | `"Player1<1><STEAM_ID><CT>" defused the bomb` |
| `bomb_begin_defuse` | Player started defusing bomb | `"Player1<1><STEAM_ID><CT>" triggered "Begin_Bomb_Defuse"` |
| `bomb_got` | Player picked up the bomb | `"Player1<1><STEAM_ID><T>" triggered "Got_The_Bomb"` |
| `bomb_dropped` | Player dropped the bomb | `"Player1<1><STEAM_ID><T>" dropped the bomb` |

### Economy Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `purchase` | Player purchased an item | `"Player1<1><STEAM_ID><CT>" purchased "m4a1"` |
| `money_change` | Player's money changed | `"Player1<1><STEAM_ID><CT>" money change +300 (total: 4500)` |

### Item Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `picked_up` | Player picked up an item | `"Player1<1><STEAM_ID><CT>" picked up "ak47"` |
| `dropped` | Player dropped an item | `"Player1<1><STEAM_ID><CT>" dropped "m4a1"` |

### Grenade Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `grenade_thrown` | Player threw a grenade | `"Player1<1><STEAM_ID><T>" threw smokegrenade` |
| `blinded` | Player was blinded by flashbang | `"Player1<1><STEAM_ID><CT>" blinded by "Player2<2><STEAM_ID><T>"` |
| `projectile_spawned` | Projectile entity spawned | `Molotov projectile spawned at coordinates` |

---

## Extended Events (Custom Classification)

These events are classified from CS2's "Unknown" type logs using pattern matching.

### Buy Zone Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `left_buyzone` | Player left the buy zone with equipment | `"Player1<1><STEAM_ID><CT>" left buyzone with [ weapon_knife weapon_usp_silencer kevlar(100) ]` |

**Use Case**: Track player economy and loadout choices at round start.

### Validation Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `userid_validated` | Steam user ID validated by server | `"Player1<1><STEAM_ID><>" STEAM USERID validated` |

**Use Case**: Confirm player authentication and track connection security.

### Achievement/Award Events

| Event Type | Description | Pattern |
|------------|-------------|---------|
| `accolade_final_3k` | Triple kill achievement | `ACCOLADE, FINAL: {3k}, Player1<1>, VALUE: 2.000000` |
| `accolade_final_4k` | Quadruple kill achievement | `ACCOLADE, FINAL: {4k}, Player1<1>, VALUE: 1.000000` |
| `accolade_final_5k` | Ace (5 kills) achievement | `ACCOLADE, FINAL: {5k}, Player1<1>, VALUE: 1.000000` |
| `accolade_final_uniqueweaponkills` | Kills with different weapons | `ACCOLADE, FINAL: {uniqueweaponkills}, Player1<1>, VALUE: 15.000000` |
| `accolade_final_taserkills` | Zeus/Taser kills | `ACCOLADE, FINAL: {taserkills}, Player1<1>, VALUE: 1.000000` |
| `accolade_final_mvp` | MVP award | `ACCOLADE, FINAL: {mvp}, Player1<1>, VALUE: 5.000000` |

**Use Case**: Track player performance, create achievement systems, and identify standout performances.

### Match Status Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `match_status_score` | Current match score update | `MatchStatus: Score: 17:19 on map "de_dust2" RoundsPlayed: 36` |
| `match_status_teams` | Team assignments | `MatchStatus: Team playing "TERRORIST": team_SHESKY` |

**Use Case**: Track match progress, implement live scoreboards, and monitor match state.

### Pause Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `match_pause_enabled` | Match pause activated | `Match pause is enabled - mp_pause_match` |
| `match_pause_disabled` | Match pause deactivated | `Match pause is disabled - TimeOutTs` |
| `match_unpause` | Match resumed from pause | `Match unpaused` |

**Use Case**: Track tactical timeouts, technical pauses, and match flow interruptions.

### Debug Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `throw_debug_molotov` | Molotov throw debug data | `"Player1" sv_throw_molotov -1943.109 1620.291 94.267 ...` |
| `throw_debug_smoke` | Smoke grenade throw debug | `"Player1" sv_throw_smokegrenade -465.938 1875.818 ...` |
| `throw_debug_flash` | Flashbang throw debug | `"Player1" sv_throw_flashgrenade -271.017 1354.833 ...` |
| `throw_debug_he` | HE grenade throw debug | `"Player1" sv_throw_hegrenade 123.456 789.012 ...` |
| `throw_debug` | Generic grenade throw debug | `"Player1" sv_throw ...` |

**Use Case**: Analyze grenade trajectories, create tactical analysis tools, and debug server-side physics.

### Server Configuration Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `server_cvar` | Server variable changed | `server_cvar: "mp_freezetime" "20"` |
| `cvar_maxrounds` | Max rounds setting | `"mp_maxrounds" = "24"` |
| `cvar_overtime` | Overtime configuration | `"mp_overtime_enable" = "1"` |
| `cvar_freezetime` | Freeze time duration | `"mp_freezetime" = "15"` |
| `cvar_tournament` | Tournament mode setting | `"mp_tournament" = "1"` |
| `cvar_mp_setting` | Other MP settings | `"mp_winlimit" = "0"` |

**Use Case**: Track server configuration changes, ensure competitive integrity, and audit match settings.

### Chat Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `chat_pause_command` | Player requested pause | `"Player1" say ".pause"` or `".forcepause"` |
| `chat_restore_command` | Player requested restore | `"Player1" say ".restore 35"` |
| `chat_ready_command` | Player signaled ready | `"Player1" say ".ready"` or `".rdy"` |
| `chat_gg` | Good game message | `"Player1" say "gg"` or `"gg wp"` |
| `chat_message` | General chat message | `"Player1" say "nice shot"` |

**Use Case**: Track match administration commands, player communication, and sportsmanship.

### Game State Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `game_over_competitive` | Competitive match ended | `Game Over: competitive de_dust2 score 17:19 after 49 min` |
| `game_over_casual` | Casual match ended | `Game Over: casual de_mirage score 8:5 after 20 min` |
| `game_over` | Generic game over | `Game Over: score 16:14` |
| `freeze_period_start` | Freeze period started | `Starting Freeze period` |

**Use Case**: Track match conclusions, calculate match duration, and trigger end-game statistics.

### Map Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `loading_map` | Server loading new map | `Loading map "de_mirage"` |
| `started_map` | Map fully loaded | `Started map "de_dust2"` |

**Use Case**: Track map rotations, server transitions, and loading times.

### Team Events (Extended)

| Event Type | Description | Example |
|------------|-------------|---------|
| `team_playing` | Team name assignment | `Team playing "TERRORIST": team_xHaPPy_` |
| `team_ct_scored` | CT team scored | `CT scored "16" points` |
| `team_t_scored` | T team scored | `TERRORIST scored "14" points` |

**Use Case**: Track team names in tournaments, monitor score progression.

### Log File Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `log_file_started` | Log recording started | `Log file started (file "logs/2025_08_19.log")` |
| `log_file_closed` | Log recording stopped | `Log file closed` |

**Use Case**: Track logging sessions, ensure data integrity, and manage log rotation.

### RCON Events

| Event Type | Description | Example |
|------------|-------------|---------|
| `rcon_command` | Remote console command | `rcon from "192.168.1.100:12345": command "mp_pause_match 1"` |

**Use Case**: Audit administrative actions, track remote management, and ensure security.

### Triggered Events (Dynamic)

| Event Type | Description | Example |
|------------|-------------|---------|
| `trigger_warmup-start` | Warmup period started | `World triggered "Warmup_Start"` |
| `trigger_match-reloaded` | Match state reloaded | `World triggered "Match_Reloaded" on "de_dust2"` |
| `trigger_sfui-notice-*` | UI notifications | `World triggered "SFUI_Notice_Round_Draw"` |
| `triggered_event` | Generic triggered event | `World triggered "Custom_Event"` |

**Use Case**: Track special game states, handle custom server events, and monitor match flow.

### Bomb Events (Extended)

| Event Type | Description | Example |
|------------|-------------|---------|
| `bomb_begin_plant` | Bomb planting started (trigger) | `"Player1" triggered "Bomb_Begin_Plant" at bombsite B` |
| `bomb_planted_trigger` | Bomb planted (trigger event) | `"Player1" triggered "Bomb_Planted"` |
| `bomb_defused_trigger` | Bomb defused (trigger event) | `"Player1" triggered "Bomb_Defused"` |

**Use Case**: Track bomb-related events not caught by main parser, analyze site preferences.

### Unknown Events

| Event Type | Description | When Used |
|------------|-------------|----------|
| `unknown` | Recognized as Unknown by cs2-log but no content | Empty or malformed Unknown type |
| `unknown_other` | Cannot be classified | Doesn't match any known pattern |

---

## Event Data Structure

Each parsed event contains:
- **event_type**: String identifier from this documentation
- **event_data**: JSON object with event-specific fields
- **timestamp**: When the event occurred
- **server_id**: Which server generated the event
- **raw_content**: Original log line

## Usage Examples

### Filtering Events by Type
```sql
-- Get all kills in the last hour
SELECT * FROM parsed_logs 
WHERE event_type = 'kill' 
AND created_at > NOW() - INTERVAL '1 hour';

-- Get all chat commands
SELECT * FROM parsed_logs 
WHERE event_type LIKE 'chat_%command';

-- Get all bomb-related events
SELECT * FROM parsed_logs 
WHERE event_type LIKE '%bomb%';
```

### Analyzing Player Performance
```sql
-- Count kills per player
SELECT 
    event_data->>'attacker' as player,
    COUNT(*) as kills
FROM parsed_logs 
WHERE event_type = 'kill'
GROUP BY player;

-- Find all achievements
SELECT * FROM parsed_logs 
WHERE event_type LIKE 'accolade%';
```

### Tracking Match Flow
```sql
-- Get match timeline
SELECT 
    event_type,
    event_data,
    created_at
FROM parsed_logs 
WHERE event_type IN (
    'match_start', 'round_start', 'round_end', 
    'match_pause_enabled', 'match_unpause', 'game_over_competitive'
)
ORDER BY created_at;
```

## Notes for Developers

1. **Event Classification Priority**: The parser checks events in a specific order. More specific patterns are checked before generic ones.

2. **Custom Events**: Servers can generate custom events. These will typically appear as `trigger_*` events if they use the trigger format.

3. **Performance Considerations**: The most common events (kills, attacks, chat) are checked first for optimal performance.

4. **Extensibility**: New event types can be added by updating the `classifyUnknownEvent` function in `parser_service.go`.

5. **Raw Data Preservation**: Even if parsing fails or returns unknown, the raw log is always preserved for future reprocessing.

## Version History

- **v1.0** - Initial event type documentation
- **v1.1** - Added extended classification for Unknown events
- **v1.2** - Enhanced categorization with 40+ event types

---

*Last Updated: August 19, 2025*