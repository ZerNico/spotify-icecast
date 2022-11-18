#!/bin/sh

curl -X POST http://localhost:8080/event --silent -o /dev/null -H "Content-Type: application/json" -d '{ "player_event":"'"$PLAYER_EVENT"'", "track_id":"'"$TRACK_ID"'", "old_track_id":"'"$OLD_TRACK_ID"'" }'