#!/usr/bin/bash
export APP_NAME=realy
export BINARY=true
export LISTEN=0.0.0.0
export PORT=3334
export PPROF=false
export SUPERUSER=npub1fjqqy4a93z5zsjwsfxqhc2764kvykfdyttvldkkkdera8dr78vhsmmleku

until false; do
    echo "Respawning.." >&2
    sleep 1
	reset
    go run ./cmd/realy/.
done
