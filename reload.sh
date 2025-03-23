#!/usr/bin/bash
until go run ./cmd/realy/.; do
    echo "Respawning.." >&2
    sleep 1
done
