package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "service-manager-read-mcp: MCP server not wired yet")
	os.Exit(0)
}
