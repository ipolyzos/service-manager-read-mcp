package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ipolyzos/service-manager-read-mcp/internal/servicemcp"
	"github.com/ipolyzos/service-manager-read-mcp/internal/version"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(version.Summary())
		return
	}

	ctx := context.Background()
	if err := servicemcp.RunStdio(ctx, os.Stderr, servicemcp.Options{}); err != nil && err != context.Canceled {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
