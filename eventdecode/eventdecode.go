package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"log"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

//go:embed events.json
var eventsJSON string

func main() {
	// Create accumulator
	var acc anthropic.Message

	// Read and process each line
	scanner := bufio.NewScanner(strings.NewReader(eventsJSON))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()

		// Parse the event
		var event anthropic.MessageStreamEventUnion
		if err := json.Unmarshal(line, &event); err != nil {
			log.Fatalf("line %d: failed to unmarshal event: %v\nLine content: %s", lineNum, err, string(line))
		}

		// Accumulate the event
		if err := acc.Accumulate(event); err != nil {
			log.Fatalf("line %d: failed to accumulate event. error from Anthropic SDK:\n%v", lineNum, err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	log.Printf("successfully accumulated %d events into message with %d content blocks", lineNum, len(acc.Content))
}
