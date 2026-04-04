package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ipolyzos/service-manager-read-mcp/internal/servicemcp"
)

func main() {
	ctx := context.Background()
	if err := servicemcp.RunStdio(ctx, os.Stderr, servicemcp.Options{}); err != nil && err != context.Canceled {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
