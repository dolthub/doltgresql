// Copyright 2023 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/dolthub/dolt/go/libraries/utils/config"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dolthub/dolt/go/cmd/dolt/cli"
	"github.com/dolthub/dolt/go/cmd/dolt/commands"
	"github.com/dolthub/dolt/go/cmd/dolt/commands/sqlserver"
	eventsapi "github.com/dolthub/dolt/go/gen/proto/dolt/services/eventsapi/v1alpha1"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/events"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/dolt/go/store/nbs"
	"github.com/dolthub/doltgresql/server"
	"github.com/fatih/color"
)

func init() {
	events.Application = eventsapi.AppID_APP_DOLTGRES

	if os.Getenv("DOLT_VERBOSE_ASSERT_TABLE_FILES_CLOSED") == "" {
		nbs.TableIndexGCFinalizerWithStackTrace = false
	}
}

const (
	chdirParam        = "chdir"
	stdInParam        = "stdin"
	stdOutParam       = "stdout"
	stdErrParam       = "stderr"
	stdOutAndErrParam = "out-and-err"

	configParam = "config"
)

func parseArgs() (flags map[string]*bool, params map[string]*string) {
	flags = make(map[string]*bool)
	params = make(map[string]*string)

	params[chdirParam] = flag.String(chdirParam, "", "set the working directory for instancemgr")
	params[stdInParam] = flag.String(stdInParam, "", "directory where applications are installed. This is the directory where subdirectories for the dolt and doltgress applications are located")
	params[stdOutParam] = flag.String(stdOutParam, "", "path where logs are stored")
	params[stdErrParam] = flag.String(stdErrParam, "", "path where systemd services are installed")
	params[stdOutAndErrParam] = flag.String(stdOutAndErrParam, "", "if using the cloudwatch agent, the directory where it is installed")
	params[configParam] = flag.String(configParam, "config.yaml", "Where the scraped metrics should be sent. Valid values are 'null', or 'cloudwatch'")

	flag.Parse()

	return flags, params
}

func main() {
	ctx := context.Background()
	args := os.Args[1:]

	_, params := parseArgs()

	err := redirectStdio(params)
	if err != nil {
		cli.PrintErrln(err.Error())
		os.Exit(1)
	}

	restoreIO := cli.InitIO()
	defer restoreIO()

	warnIfMaxFilesTooLow()

	fs := filesys.LocalFS
	dEnv := env.Load(ctx, env.GetCurrentUserHomeDir, fs, doltdb.LocalDirDoltDB, server.Version)

	args, err = configureDataDir(args)
	if err != nil {
		cli.PrintErrln(err.Error())
		os.Exit(1)
	}

	config, err := server.ReadConfigFromYamlFile(*params[configParam])
	if err != nil {
		cli.PrintErrln(err.Error())
		os.Exit(1)
	}

	err = runServer(ctx, dEnv, config)
	if err != nil {
		cli.PrintErrln(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

// configureDataDir sets the --data-dir argument as appropriate if it isn't specified
func configureDataDir(args []string) (outArgs []string, err error) {
	// We can't use the argument parser yet since it relies on the environment, so we'll handle the data directory
	// argument here. This will also remove it from the args, so that the Dolt layer doesn't try to move the directory
	// again (in the case of relative paths).
	var hasDataDirArgument bool
	for i, arg := range args {
		arg = strings.ToLower(arg)
		if arg == "--data-dir" {
			if len(args) <= i+1 {
				return args, fmt.Errorf("--data-dir is missing the directory")
			}
			hasDataDirArgument = true
			break
		} else if strings.HasPrefix(arg, "--data-dir=") {
			hasDataDirArgument = true
		}
	}

	if hasDataDirArgument {
		return args, nil
	}

	// We should use the directory as pointed to by "DOLTGRES_DATA_DIR", if has been set, otherwise we'll use the default
	var dbDir string
	if envDir := os.Getenv(server.DOLTGRES_DATA_DIR); len(envDir) > 0 {
		dbDir = envDir
		fileInfo, err := os.Stat(dbDir)
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dbDir, 0755); err != nil {
				return args, err
			}
		} else if err != nil {
			return args, err
		} else if !fileInfo.IsDir() {
			return args, fmt.Errorf("Attempted to use the directory `%s` as the DoltgreSQL database directory, "+
				"however the preceding is a file and not a directory. Please change the environment variable `%s` so "+
				"that it points to a directory.", dbDir, server.DOLTGRES_DATA_DIR)
		}
	} else {
		homeDir, err := env.GetCurrentUserHomeDir()
		if err != nil {
			return args, err
		}
		dbDir = filepath.Join(homeDir, server.DOLTGRES_DATA_DIR_DEFAULT)
		fileInfo, err := os.Stat(dbDir)
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dbDir, 0755); err != nil {
				return args, err
			}
		} else if err != nil {
			return args, err
		} else if !fileInfo.IsDir() {
			return args, fmt.Errorf("Attempted to use the directory `%s` as the DoltgreSQL database directory, "+
				"however the preceding is a file and not a directory. Please change the environment variable `%s` so "+
				"that it points to a directory.", dbDir, server.DOLTGRES_DATA_DIR)
		}
	}

	// alter the data dir argument provided to dolt arg processing
	args = append([]string{"--data-dir", dbDir}, args...)

	return args, nil
}

// runServer launches a server on the env given and waits for it to finish
func runServer(ctx context.Context, dEnv *env.DoltEnv, cfg *server.Config) error {
	// Emit a usage event in the background while we start the server.
	// Dolt is more permissive with events: it emits events even if the command fails in the earliest possible phase,
	// we emit an event only if we got far enough to attempt to launch a server (and we may not emit it if the server
	// dies quickly enough).
	//
	// We also emit a heartbeat event every 24 hours the server is running.
	// All events will be tagged with the doltgresql app id.
	go emitUsageEvent(ctx, dEnv)

	controller, err := server.RunOnDisk(ctx, cfg, dEnv)
	if err != nil {
		return err
	}

	return controller.WaitForStop()
}

func paramVal(params map[string]*string, key string) (string, bool) {
	val, ok := params[key]
	if !ok || val == nil || *val == "" {
		return "", false
	}

	return *val, true
}

func redirectStdio(params map[string]*string) error {
	// Currently goland doesn't support running with a different working directory when using go modules.
	// This is a hack that allows a different working directory to be set after the application starts using
	// chdir=<DIR>.  The syntax is not flexible and must match exactly this.
	if chdir, ok := paramVal(params, chdirParam); ok {
		err := os.Chdir(chdir)

		if err != nil {
			panic(err)
		}
	}

	if stdInFile, ok := paramVal(params, stdInParam); ok {
		cli.Println("Using file contents as stdin:", stdInFile)

		f, err := os.Open(stdInFile)
		if err != nil {
			return fmt.Errorf("Failed to open %s: %w", stdInFile, err)
		}

		os.Stdin = f
	}

	if filename, ok := paramVal(params, stdOutParam); ok {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Failed to open %s for writing: %w", filename, err)
		}
		cli.Println("Stdout being written to", filename)
		cli.CliOut = f
		color.NoColor = true
	}

	if filename, ok := paramVal(params, stdErrParam); ok {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Failed to open %s for writing: %w", filename, err)
		}
		cli.Println("Stderr being written to", filename)
		cli.CliErr = f
		color.NoColor = true
	}

	if filename, ok := paramVal(params, stdOutAndErrParam); ok {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Failed to open %s for writing: %w", filename, err)
		}
		cli.Println("Stdout and Stderr being written to", filename)
		cli.CliOut = f
		cli.CliErr = f
		color.NoColor = true
	}

	return nil
}

// emitUsageEvent emits a usage event to the event server
func emitUsageEvent(ctx context.Context, dEnv *env.DoltEnv) {
	metricsDisabled := dEnv.Config.GetStringOrDefault(config.MetricsDisabled, "false")
	disabled, err := strconv.ParseBool(metricsDisabled)
	if err != nil || disabled {
		return
	}

	emitter, closeFunc, err := commands.GRPCEmitterForConfig(dEnv)
	if err != nil {
		return
	}
	defer closeFunc()

	evt := events.NewEvent(sqlserver.SqlServerCmd{}.EventType())
	evtCollector := events.NewCollector(server.Version, emitter)
	evtCollector.CloseEventAndAdd(evt)
	clientEvents := evtCollector.Close()

	_ = emitter.LogEvents(ctx, server.Version, clientEvents)
}
