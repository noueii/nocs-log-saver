# CS2 Event Types - Quick Reference

## Most Common Events

### Combat & Gameplay
- `kill` - Player killed another
- `attack` - Player damaged another  
- `kill_assist` - Assisted in kill
- `suicide` - Self-elimination
- `killed_by_bomb` - Bomb explosion death

### Rounds & Matches
- `round_start` - Round began
- `round_end` - Round finished
- `match_start` - Match began
- `match_end` - Match finished (GameOver)
- `game_over_competitive` - Competitive match ended with details

### Player Status
- `player_connect` - Connected to server
- `player_disconnect` - Left server
- `player_entered` - Joined game
- `team_switch` - Changed teams

### Bomb Operations  
- `bomb_planted` - Bomb planted
- `bomb_defused` - Bomb defused
- `bomb_begin_defuse` - Started defusing
- `bomb_got` - Picked up bomb
- `bomb_dropped` - Dropped bomb
- `bomb_begin_plant` - Started planting

### Economy & Items
- `purchase` - Bought item
- `money_change` - Money updated
- `picked_up` - Picked up item
- `dropped` - Dropped item
- `left_buyzone` - Left buy area with loadout

### Grenades
- `grenade_thrown` - Threw grenade
- `blinded` - Flashbang blinded
- `projectile_spawned` - Grenade spawned
- `throw_debug_smoke` - Smoke trajectory debug
- `throw_debug_flash` - Flash trajectory debug
- `throw_debug_molotov` - Molotov trajectory debug
- `throw_debug_he` - HE trajectory debug

### Communication
- `chat` - Regular chat (parsed)
- `chat_message` - General chat (unknown)
- `chat_pause_command` - `.pause` command
- `chat_restore_command` - `.restore` command
- `chat_ready_command` - `.ready` command
- `chat_gg` - "gg" message

### Match Administration
- `match_pause_enabled` - Pause activated
- `match_pause_disabled` - Pause deactivated
- `match_unpause` - Resumed from pause
- `match_status_score` - Score update
- `match_status_teams` - Team assignments

### Server Configuration
- `server_cvar` - Server variable changed
- `cvar_maxrounds` - Max rounds changed
- `cvar_freezetime` - Freeze time changed
- `cvar_overtime` - Overtime settings
- `cvar_tournament` - Tournament mode

### Achievements (Accolades)
- `accolade_final_3k` - Triple kill
- `accolade_final_4k` - Quad kill
- `accolade_final_5k` - Ace
- `accolade_final_mvp` - MVP award
- `accolade_final_uniqueweaponkills` - Weapon variety
- `accolade_final_taserkills` - Zeus kills

### Map & Game State
- `loading_map` - Loading new map
- `started_map` - Map loaded
- `freeze_period_start` - Freeze time began
- `game_commencing` - Game starting
- `userid_validated` - Steam ID validated

### Special Events
- `trigger_warmup-start` - Warmup began
- `trigger_match-reloaded` - Match restored
- `trigger_sfui-notice-*` - UI notifications
- `rcon_command` - Remote admin command
- `log_file_started` - Logging started
- `log_file_closed` - Logging stopped

### Team Scoring
- `team_scored` - Team scored points
- `team_ct_scored` - CT team scored
- `team_t_scored` - T team scored
- `team_playing` - Team name assignment
- `team_notice` - Team notification

## Event Type Patterns

| Pattern | Description | Example Types |
|---------|-------------|---------------|
| `chat_*` | Chat-related events | `chat_message`, `chat_pause_command` |
| `bomb_*` | Bomb-related events | `bomb_planted`, `bomb_defused` |
| `accolade_*` | Achievement events | `accolade_final_3k`, `accolade_final_mvp` |
| `cvar_*` | Server settings | `cvar_maxrounds`, `cvar_freezetime` |
| `throw_debug_*` | Grenade debug info | `throw_debug_smoke`, `throw_debug_flash` |
| `match_*` | Match state events | `match_start`, `match_pause_enabled` |
| `trigger_*` | Triggered game events | `trigger_warmup-start` |
| `player_*` | Player state changes | `player_connect`, `player_entered` |
| `team_*` | Team-related events | `team_switch`, `team_scored` |
| `game_over_*` | Match end events | `game_over_competitive` |

## SQL Quick Queries

```sql
-- Get event type distribution
SELECT event_type, COUNT(*) as count 
FROM parsed_logs 
GROUP BY event_type 
ORDER BY count DESC;

-- Find all combat events
SELECT * FROM parsed_logs 
WHERE event_type IN ('kill', 'attack', 'kill_assist');

-- Get match timeline
SELECT event_type, created_at 
FROM parsed_logs 
WHERE event_type LIKE '%match%' OR event_type LIKE '%round%'
ORDER BY created_at;

-- Find all unknown events
SELECT * FROM parsed_logs 
WHERE event_type LIKE 'unknown%';
```

## API Endpoints

```bash
# Get all event types with counts
GET /api/event-types?server_id=server1

# Get logs filtered by event type
GET /api/logs?type=parsed&event_type=kill

# Test parsing with specific logs
POST /api/parse-test
Body: {"logs": "L 08/19/2025 - 19:03:31: ..."}
```

---

*For full documentation with examples and use cases, see [CS2_EVENT_TYPES.md](./CS2_EVENT_TYPES.md)*