package main

import (
	"context"
	"fmt"
	"log"
	"os"

	builder "github.com/dolthub/doltgresql/utils/doltgres_builder"
)

func main() {
	commitList := os.Args[1:]
	if len(commitList) < 1 {
		helpStr := "doltgres-builder takes DoltgreSQL commit shas or tags as arguments\n" +
			"and builds corresponding binaries to a path specified\n" +
			"by DOLTGRES_BIN\n" +
			"If DOLTGRES_BIN is not set, ./doltgresBin will be used\n" +
			"usage: doltgres-builder dccba46 4bad226 ...\n" +
			"usage: doltgres-builder v0.19.0 v0.22.6 ...\n" +
			"set DEBUG=1 to run in debug mode\n"
		fmt.Print(helpStr)
		os.Exit(2)
	}

	err := builder.Run(context.Background(), commitList)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
