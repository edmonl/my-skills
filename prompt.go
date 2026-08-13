package main

import _ "embed"

// agentPrompt contains the instructions printed by the prompt subcommand.
//
//go:embed prompt.md
var agentPrompt string
