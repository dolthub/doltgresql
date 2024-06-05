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
	"os"
	"path/filepath"
	"strconv"

	"github.com/dolthub/dolt/go/cmd/dolt/cli"
	"github.com/dolthub/dolt/go/cmd/dolt/commands"
	"github.com/dolthub/dolt/go/cmd/dolt/commands/sqlserver"
	eventsapi "github.com/dolthub/dolt/go/gen/proto/dolt/services/eventsapi/v1alpha1"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/events"
	"github.com/dolthub/dolt/go/libraries/utils/config"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/dolt/go/store/nbs"
	"github.com/fatih/color"

	"github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/servercfg"
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
	configParam       = "config"
	dataDirParam   = "data-dir"
	defaultCfgFile = "config.yaml"

	versionFlag    = "version"
	configHelpFlag = "config-help"
)

func parseArgs() (flags map[string]*bool, params map[string]*string) {
	flag.Usage = func() {
		cli.Println("Usage: doltgres [options]")
		cli.Println("Options:")
		flag.PrintDefaults()
	}

	flags = make(map[string]*bool)
	params = make(map[string]*string)

	params[chdirParam] = flag.String(chdirParam, "", "set the working directory for doltgres")
	params[stdInParam] = flag.String(stdInParam, "", "file to use as stdin")
	params[stdOutParam] = flag.String(stdOutParam, "", "file to use as stdout")
	params[stdErrParam] = flag.String(stdErrParam, "", "file to use as stderr")
	params[stdOutAndErrParam] = flag.String(stdOutAndErrParam, "", "file to use as stdout and stderr")
	params[configParam] = flag.String(configParam, "", "path to the config file")
	params[dataDirParam] = flag.String(dataDirParam, "", "path to the directory where doltgres databases are stored")

	flags[versionFlag] = flag.Bool(versionFlag, false, "print the version")
	flags[configHelpFlag] = flag.Bool(configHelpFlag, false, "print the config file help")

	flag.Parse()

	return flags, params
}

func main() {
	ctx := context.Background()
	flags, params := parseArgs()

	if *flags[versionFlag] {
		cli.Println("Doltgres version", server.Version)
		os.Exit(0)
	} else if *flags[configHelpFlag] {
		cli.Println(servercfg.ConfigHelp)
		os.Exit(0)
	}

	err := redirectStdio(params)
	if err != nil {
		handleErrAndExitCode(err)
	}

	restoreIO := cli.InitIO()
	defer restoreIO()

	warnIfMaxFilesTooLow()

	fs := filesys.LocalFS
	// dEnv will be reloaded at server start to point to the data dir on the server config
	dEnv := env.Load(ctx, env.GetCurrentUserHomeDir, fs, doltdb.LocalDirDoltDB, server.Version)
	
	cfg, err := loadServerConfig(params, fs)
	if err != nil {
		handleErrAndExitCode(err)
	}

	err = setupDataDir(params, cfg, fs)
	if err != nil {
		handleErrAndExitCode(err)
	}
	
	// TODO: override other aspects of cfg with command line params
	
	err = runServer(ctx, dEnv, cfg)
	handleErrAndExitCode(err)
}

func handleErrAndExitCode(err error) {
	if err != nil {
		cli.PrintErrln(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

// setupDataDir sets the appropriate data dir for the config given based on parameters and env, and creates the data
// dir as necessary.
func setupDataDir(params map[string]*string, cfg *servercfg.DoltgresConfig, fs filesys.Filesys) error {
	dataDir, dataDirType, err := getDataDirFromParams(params)
	if err != nil {
		return err
	}

	// determine the data dir to use, in order of preference: 1) explicit flag, 2) env var, 3) default
	if dataDirType == dataDirExplicitParam {
		cfg.DataDirStr = &dataDir
	} else if dataDirType == dataDirEnv && (cfg.DataDirStr == nil || *cfg.DataDirStr == servercfg.DefaultDataDir) {
		cfg.DataDirStr = &dataDir
	} else {
		def := servercfg.DefaultDataDir
		cfg.DataDirStr = &def
	}

	dataDirPath := cfg.DataDir()
	dataDirExists, isDir := fs.Exists(dataDirPath)
	if !dataDirExists {
		if err := fs.MkDirs(dataDirPath); err != nil {
			return fmt.Errorf("failed to make dir '%s': %w", dataDirPath, err)
		}
	} else if !isDir {
		return fmt.Errorf("cannot use file %s as doltgres data directory", dataDirPath)
	}
	
	return nil
}

// loadServerConfig loads server configuration in the following order:
// 1. If the --config flag is provided, loads the config from the file at the path provided, or returns an errors if it cannot.
// 2. If the default config file config.yaml exists, attempts to load it, but doesn't return an error if it doesn't exist.
// 3. If neither of the above are successful, returns the default config server config.
func loadServerConfig(params map[string]*string, fs filesys.Filesys) (*servercfg.DoltgresConfig, error) {
	configFilePath, configFilePathSpecified := paramVal(params, configParam)
	
	if configFilePathSpecified {
		cfgPathExists, isDir := fs.Exists(configFilePath)
		if !cfgPathExists {
			return nil, fmt.Errorf("config file not found at %s", configFilePath)
		} else if isDir {
			return nil, fmt.Errorf("cannot use directory %s for config file", configFilePath)
		}

		return servercfg.ReadConfigFromYamlFile(fs, configFilePath)
	} else {
		cfgPathExists, isDir := fs.Exists(defaultCfgFile)
		if cfgPathExists && !isDir {
			return servercfg.ReadConfigFromYamlFile(fs, configFilePath)
		}
	}

	return servercfg.DefaultServerConfig(), nil
}

type dataDirType byte
const (
	dataDirExplicitParam dataDirType = iota
	dataDirEnv
	dataDirDefault
)

// getDataDirFromParams returns the dataDir to be used by the server, along with whether it was explicitly set.
func getDataDirFromParams(params map[string]*string) (string, dataDirType, error) {
	if dataDir, ok := paramVal(params, dataDirParam); ok {
		return dataDir, dataDirExplicitParam, nil
	}

	// We should use the directory as pointed to by "DOLTGRES_DATA_DIR", if has been set, otherwise we'll use the default
	if envDir := os.Getenv(server.DOLTGRES_DATA_DIR); len(envDir) > 0 {
		return envDir, dataDirEnv, nil
	} else {
		homeDir, err := env.GetCurrentUserHomeDir()
		if err != nil {
			return "", dataDirDefault, fmt.Errorf("failed to get current user's home directory: %w", err)
		}

		dbDir := filepath.Join(homeDir, server.DOLTGRES_DATA_DIR_DEFAULT)
		return dbDir, dataDirDefault, nil
	}
}

// runServer launches a server on the env given and waits for it to finish
func runServer(ctx context.Context, dEnv *env.DoltEnv, cfg *servercfg.DoltgresConfig) error {
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
