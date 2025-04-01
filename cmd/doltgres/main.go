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
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/cockroachdb/errors"
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
	"github.com/mitchellh/go-wordwrap"

	"github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/servercfg"
	"github.com/dolthub/doltgresql/utils"
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
	dataDirParam      = "data-dir"
	defaultCfgFile    = "config.yaml"

	profilePath  = "--prof-path"
	profFlag     = "--prof"
	cpuProf      = "cpu"
	memProf      = "mem"
	blockingProf = "blocking"
	traceProf    = "trace"

	versionFlag    = "version"
	configHelpFlag = "config-help"

	configHelpText = "Path to the config file.\n" +
		"If not provided, ./config.yaml will be used if it exists."
	dataDirHelpText = "Path to the directory where doltgres databases are stored.\n" +
		"If not provided, the value in config.yaml will be used. If that's not provided either, the value of the " +
		"DOLTGRES_DATA_DIR environment variable will be used if set. Otherwise $HOME/doltgres/databases will be used. " +
		"The directory will be created if it doesn't exist."
)

func parseArgs() (flags map[string]*bool, params map[string]*string) {
	flag.Usage = func() {
		cli.Println("Usage: doltgres [options]")
		cli.Println("Options:")
		PrintDefaults(flag.CommandLine)
	}

	flags = make(map[string]*bool)
	params = make(map[string]*string)

	params[configParam] = flag.String(configParam, "", configHelpText)
	params[dataDirParam] = flag.String(dataDirParam, "", dataDirHelpText)
	params[chdirParam] = flag.String(chdirParam, "", "set the working directory for doltgres")
	params[stdInParam] = flag.String(stdInParam, "", "file to use as stdin")
	params[stdOutParam] = flag.String(stdOutParam, "", "file to use as stdout")
	params[stdErrParam] = flag.String(stdErrParam, "", "file to use as stderr")
	params[stdOutAndErrParam] = flag.String(stdOutAndErrParam, "", "file to use as stdout and stderr")

	flags[versionFlag] = flag.Bool(versionFlag, false, "print the version")
	flags[configHelpFlag] = flag.Bool(configHelpFlag, false, "print the config file help")

	flag.Parse()

	return flags, params
}

// PrintDefaults is modified from the flag package to control the order of printing
func PrintDefaults(fs *flag.FlagSet) {
	helpOrder := []string{
		configParam,
		dataDirParam,
		configHelpFlag,
		versionFlag,
		chdirParam,
		stdInParam,
		stdOutParam,
		stdErrParam,
		stdOutAndErrParam,
	}

	for _, fName := range helpOrder {
		f := fs.Lookup(fName)
		var b strings.Builder
		fmt.Fprintf(&b, "  -%s", f.Name) // Two spaces before -; see next two comments.
		name, usage := flag.UnquoteUsage(f)
		if len(name) > 0 {
			b.WriteString(" ")
			b.WriteString(name)
		}
		// Boolean flags of one ASCII letter are so common we
		// treat them specially, putting their usage on the same line.
		if b.Len() <= 4 { // space, space, '-', 'x'.
			b.WriteString("\t")
		} else {
			// Four spaces before the tab triggers good alignment
			// for both 4- and 8-space tab stops.
			b.WriteString("\n    \t")
		}
		// wrap at 80 chars
		usage = wordwrap.WrapString(usage, 76)
		b.WriteString(strings.ReplaceAll(usage, "\n", "\n    \t"))

		if f.DefValue != "false" && f.DefValue != "" {
			fmt.Fprintf(&b, " (default %v)", f.DefValue)
		}
		fmt.Fprint(fs.Output(), b.String(), "\n")
	}
}

func main() {
	args := os.Args[1:]

	if len(args) >= 2 {
		profilingOptions := utils.ProfilingOptions{}
		doneDebugFlags := false
		for !doneDebugFlags && len(args) > 0 {
			switch args[0] {
			case profilePath:
				profilingOptions.Path = args[1]
				if _, err := os.Stat(profilingOptions.Path); errors.Is(err, os.ErrNotExist) {
					panic(errors.Errorf("profile path does not exist: %s", profilingOptions.Path))
				}
			case profFlag:
				switch args[1] {
				case cpuProf:
					profilingOptions.CPU = true
				case memProf:
					profilingOptions.Memory = true
				case blockingProf:
					profilingOptions.Block = true
				case traceProf:
					profilingOptions.Trace = true
				default:
					panic("Unexpected prof flag: " + args[1])
				}
			default:
				doneDebugFlags = true
			}

			args = args[2:]
		}

		if profilingOptions.HasOptions() {
			utils.StartProfiling(profilingOptions)
			defer utils.StopProfiling()
		}

		os.Args = append([]string{os.Args[0]}, args...)
	}
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

	cfg, loadedFromDisk, err := loadServerConfig(params, fs)
	if err != nil {
		handleErrAndExitCode(err)
	}

	err = setupDataDir(params, cfg, loadedFromDisk, fs)
	if err != nil {
		handleErrAndExitCode(err)
	}

	// TODO: override other aspects of cfg with command line params

	// This allows catching SIGTERM when server is stopped.
	// It causes server.Close() to be called.
	var stop context.CancelFunc
	ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = runServer(ctx, dEnv, cfg)
	handleErrAndExitCode(err)
}

func handleErrAndExitCode(err error) {
	utils.StopProfiling()
	if err != nil {
		cli.PrintErrln(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

// setupDataDir sets the appropriate data dir for the config given based on parameters and env, and creates the data
// dir as necessary.
func setupDataDir(params map[string]*string, cfg *servercfg.DoltgresConfig, cfgLoadedFromDisk bool, fs filesys.Filesys) error {
	dataDir, dataDirType, err := getDataDirFromParams(params)
	if err != nil {
		return err
	}

	// This logic chooses a data dir in the following order of preference:
	// 1) explicit flag
	// 2) env var if no config file, or if the config file doesn't have a data dir
	// 3) default value if no config file, or if the config file doesn't have a data dir
	// 4) the value in the config file
	if dataDirType == dataDirExplicitParam || !cfgLoadedFromDisk || cfg.DataDirStr == nil {
		cfg.DataDirStr = &dataDir
	}

	dataDirPath := cfg.DataDir()
	dataDirExists, isDir := fs.Exists(dataDirPath)
	if !dataDirExists {
		if err := fs.MkDirs(dataDirPath); err != nil {
			return errors.Errorf("failed to make dir '%s': %w", dataDirPath, err)
		}
	} else if !isDir {
		return errors.Errorf("cannot use file %s as doltgres data directory", dataDirPath)
	}

	return nil
}

// loadServerConfig loads server configuration in the following order:
// 1. If the --config flag is provided, loads the config from the file at the path provided, or returns an errors if it cannot.
// 2. If the default config file config.yaml exists, attempts to load it, but doesn't return an error if it doesn't exist.
// 3. If neither of the above are successful, returns the default config server config.
// The second result param is a boolean indicating whether a file was loaded, since we vary later initialization
// behavior depending on whether we loaded a config file from disk.
func loadServerConfig(params map[string]*string, fs filesys.Filesys) (*servercfg.DoltgresConfig, bool, error) {
	configFilePath, configFilePathSpecified := paramVal(params, configParam)

	if configFilePathSpecified {
		cfgPathExists, isDir := fs.Exists(configFilePath)
		if !cfgPathExists {
			return nil, false, errors.Errorf("config file not found at %s", configFilePath)
		} else if isDir {
			return nil, false, errors.Errorf("cannot use directory %s for config file", configFilePath)
		}

		cfg, err := servercfg.ReadConfigFromYamlFile(fs, configFilePath)
		return cfg, true, err
	} else {
		cfgPathExists, isDir := fs.Exists(defaultCfgFile)
		if cfgPathExists && !isDir {
			cfg, err := servercfg.ReadConfigFromYamlFile(fs, defaultCfgFile)
			return cfg, true, err
		}
	}

	return servercfg.DefaultServerConfig(), false, nil
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

	if envDir := os.Getenv(servercfg.DOLTGRES_DATA_DIR); len(envDir) > 0 {
		return envDir, dataDirEnv, nil
	} else {
		homeDir, err := env.GetCurrentUserHomeDir()
		if err != nil {
			return "", dataDirDefault, errors.Errorf("failed to get current user's home directory: %w", err)
		}

		dbDir := filepath.Join(homeDir, servercfg.DOLTGRES_DATA_DIR_DEFAULT)
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
			return errors.Errorf("Failed to open %s: %w", stdInFile, err)
		}

		os.Stdin = f
	}

	if filename, ok := paramVal(params, stdOutParam); ok {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return errors.Errorf("Failed to open %s for writing: %w", filename, err)
		}
		cli.Println("Stdout being written to", filename)
		cli.CliOut = f
		color.NoColor = true
	}

	if filename, ok := paramVal(params, stdErrParam); ok {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return errors.Errorf("Failed to open %s for writing: %w", filename, err)
		}
		cli.Println("Stderr being written to", filename)
		cli.CliErr = f
		color.NoColor = true
	}

	if filename, ok := paramVal(params, stdOutAndErrParam); ok {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return errors.Errorf("Failed to open %s for writing: %w", filename, err)
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
