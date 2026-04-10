package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"dev_tool/internal/app/dtool/memory"
)

func main() {
	source := flag.String(`source`, ``, `legacy sqlite memory db path`)
	target := flag.String(`target`, ``, `memory root directory for markdown fragments`)
	flag.Parse()

	report, err := memory.MigrateLegacyDB(context.Background(), *source, *target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate memory fragments failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("migrated active=%d trash=%d files=%d\n", report.ActiveCount, report.TrashCount, len(report.Files))
}
