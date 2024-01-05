package doltgres_builder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"

	builder "github.com/dolthub/dolt/go/performance/utils/dolt_builder"
	"golang.org/x/sync/errgroup"
)

const envDoltgresBin = "DOLTGRES_BIN"

func Run(parentCtx context.Context, commitList []string) error {
	doltgresBin, err := getDoltgresBin()
	if err != nil {
		return err
	}

	// check for git on path
	err = builder.GitVersion(parentCtx)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// make temp dir for cloning/copying doltgres source
	tempDir := filepath.Join(cwd, "clones-copies")
	err = os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		return err
	}

	// clone doltgres source
	err = builder.GitCloneBare(parentCtx, tempDir, GithubDoltgres)
	if err != nil {
		return err
	}

	repoDir := filepath.Join(tempDir, "doltgresql.git")

	withKeyCtx, cancel := context.WithCancel(parentCtx)
	g, ctx := errgroup.WithContext(withKeyCtx)

	// handle user interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		<-quit
		defer wg.Done()
		signal.Stop(quit)
		cancel()
	}()

	for _, commit := range commitList {
		commit := commit // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			return buildBinaries(ctx, tempDir, repoDir, doltgresBin, commit)
		})
	}

	builderr := g.Wait()
	close(quit)
	wg.Wait()

	// remove clones-copies after all go routines complete
	// will exit successfully if removal fails
	if err := os.RemoveAll(tempDir); err != nil {
		fmt.Printf("WARN: %s was not removed\n", tempDir)
		fmt.Printf("WARN: error: %v\n", err)
	}

	if builderr != nil {
		return builderr
	}

	return nil
}

// getDoltgresBin creates and returns the absolute path for DOLTGRES_BIN
// if it was found, otherwise uses the current working directory
// as the parent directory for a `doltgresBin` directory
func getDoltgresBin() (string, error) {
	var doltgresBin string
	dir := os.Getenv(envDoltgresBin)
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		doltgresBin = filepath.Join(cwd, "doltgresBin")
	} else {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return "", err
		}
		doltgresBin = abs
	}
	err := os.MkdirAll(doltgresBin, os.ModePerm)
	if err != nil {
		return "", err
	}
	return doltgresBin, nil
}

// buildBinaries builds a doltgres binary at the given commit and stores it in the doltgresBin
func buildBinaries(ctx context.Context, tempDir, repoDir, doltgresBinDir, commit string) error {
	checkoutDir := filepath.Join(tempDir, commit)
	if err := os.MkdirAll(checkoutDir, os.ModePerm); err != nil {
		return err
	}

	err := builder.GitCheckoutTree(ctx, repoDir, checkoutDir, commit)
	if err != nil {
		return err
	}

	commitDir := filepath.Join(doltgresBinDir, commit)
	if err := os.MkdirAll(commitDir, os.ModePerm); err != nil {
		return err
	}

	parserScriptPath := filepath.Join(checkoutDir, "postgres", "parser", "build.sh")

	command, err := goBuild(ctx, parserScriptPath, checkoutDir, commitDir)
	if err != nil {
		return err
	}

	return doltgresVersion(ctx, commitDir, command)
}

// goBuild builds the doltgres parser and doltgres binary and returns the filename for the doltgres binary
func goBuild(ctx context.Context, parserScriptPath, source, dest string) (string, error) {
	buildParser := builder.ExecCommand(ctx, "/bin/bash", "-c", parserScriptPath)
	err := buildParser.Run()
	if err != nil {
		return "", err
	}

	doltFileName := "doltgres"
	if runtime.GOOS == "windows" {
		doltFileName = "doltgres.exe"
	}
	toBuild := filepath.Join(dest, doltFileName)
	build := builder.ExecCommand(ctx, "go", "build", "-o", toBuild, ".")
	build.Dir = source
	err = build.Run()
	if err != nil {
		return "", err
	}
	return toBuild, nil
}

// doltgresVersion prints doltgres version of binary
func doltgresVersion(ctx context.Context, dir, command string) error {
	doltgresVersion := builder.ExecCommand(ctx, command, "version")
	doltgresVersion.Stderr = os.Stderr
	doltgresVersion.Stdout = os.Stdout
	doltgresVersion.Dir = dir
	return doltgresVersion.Run()
}
