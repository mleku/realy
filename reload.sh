#!/usr/bin/bash
until false; do
    echo "Respawning.." >&2
    sleep 1
    go run ./cmd/realy/.
done
