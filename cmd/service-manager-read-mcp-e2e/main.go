package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ipolyzos/service-manager-read-mcp/internal/servicemcp"
	"github.com/ipolyzos/service-manager-read-mcp/internal/servicemcp/e2etest"
)

func main() {
	ctx := context.Background()
	if err := servicemcp.RunStdio(ctx, os.Stderr, e2etest.StdioOptions()); err != nil && err != context.Canceled {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
