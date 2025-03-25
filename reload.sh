#!/usr/bin/bash
until false; do
    echo "Respawning.." >&2
    sleep 1
	reset
    go run ./cmd/realy/.
done
