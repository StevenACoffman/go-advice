#!/bin/bash
go run . \
    /Users/steve/Documents/agent-orange/go-advice/Sources/benbjohnson \
    /Users/steve/Documents/agent-orange/go-advice/ai-skill/distill_step
cd distill_step

find $(pwd) -maxdepth 1 -type f -not -path '*/\.*' | sort