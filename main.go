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
	"fmt"
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
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/events"
	"github.com/dolthub/dolt/go/libraries/utils/argparser"
	"github.com/dolthub/dolt/go/libraries/utils/config"
	"github.com/dolthub/dolt/go/libraries/utils/file"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/dolt/go/store/nbs"
	"github.com/dolthub/dolt/go/store/util/tempfiles"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/fatih/color"
	"github.com/tidwall/gjson"

	"github.com/dolthub/doltgresql/server"
)

var doltgresCommands = cli.NewSubCommandHandler("doltgresql", "it's git for data", []cli.Command{
	commands.InitCmd{},
	commands.ConfigCmd{},
	commands.VersionCmd{VersionStr: server.Version},
	sqlserver.SqlServerCmd{VersionStr: server.Version},
})
var globalArgParser = cli.CreateGlobalArgParser("doltgresql")

func init() {
	events.Application = eventsapi.AppID_APP_DOLTGRES

	if os.Getenv("DOLT_VERBOSE_ASSERT_TABLE_FILES_CLOSED") == "" {
		nbs.TableIndexGCFinalizerWithStackTrace = false
	}
}

const (
	chdirFlag        = "--chdir"
	stdInFlag        = "--stdin"
	stdOutFlag       = "--stdout"
	stdErrFlag       = "--stderr"
	stdOutAndErrFlag = "--out-and-err"
)

func main() {
	ctx := context.Background()

	args := os.Args[1:]

	args, err := redirectStdio(args)
	if err != nil {
		cli.PrintErrln(err.Error())
		os.Exit(1)
	}

	restoreIO := cli.InitIO()
	defer restoreIO()

	warnIfMaxFilesTooLow()
	
	fs := filesys.LocalFS
	dEnv := env.Load(ctx, env.GetCurrentUserHomeDir, fs, doltdb.LocalDirDoltDB, server.Version)

	globalConfig, _ := dEnv.Config.GetConfig(env.GlobalConfig)

	// Inject the "sql-server" command if no other commands were given
	if len(args) == 0 || (len(args) > 0 && strings.HasPrefix(args[0], "-")) {
		args = append([]string{"sql-server"}, args...)
	}

	args, err = configureDataDir(args)
	if err != nil {
		cli.PrintErrln(err.Error())
		os.Exit(1)
	}

	apr, args, subCommandName, err := parseGlobalArgsAndSubCommandName(globalConfig, args)
	if err != nil {
		cli.PrintErrln(err.Error())
		os.Exit(1)
	}

	// The sql-server command has special cased logic since it doesn't invoke a Dolt command directly, but runs a server
	// and waits for it to finish
	if subCommandName == "sql-server" {
		err = runServer(ctx, dEnv, args[1:])
		if err != nil {	
			cli.PrintErrln(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Otherwise, attempt to run the command indicated
	cliCtx, err := configureCliCtx(subCommandName, apr, fs, dEnv, ctx)
	if err != nil {
		cli.PrintErrln(err.Error())
		os.Exit(1)
	}

	exitCode := doltgresCommands.Exec(ctx, "doltgresql", args, dEnv, cliCtx)
	os.Exit(exitCode)
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

func configureCliCtx(subcommand string, apr *argparser.ArgParseResults, fs filesys.Filesys, dEnv *env.DoltEnv, ctx context.Context) (cli.CliContext, error) {
	dataDir, hasDataDir := apr.GetValue(commands.DataDirFlag)
	if hasDataDir {
		// If a relative path was provided, this ensures we have an absolute path everywhere.
		dataDir, err := fs.Abs(dataDir)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for %s: %w", dataDir, err)
		}
		if ok, dir := fs.Exists(dataDir); !ok || !dir {
			return nil, fmt.Errorf("data directory %s does not exist", dataDir)
		}
	}

	if dEnv.CfgLoadErr != nil {
		return nil, fmt.Errorf("failed to load the global config: %w", dEnv.CfgLoadErr)
	}

	if dEnv.HasDoltDataDir() {
		cwd, _ := dEnv.FS.Abs(".")
		return nil, fmt.Errorf("Cannot start a server within a directory containing a Dolt or Doltgres database. "+
			"To use the current directory (%s) as a database, start the server from the parent directory.", cwd)
	}

	err := reconfigIfTempFileMoveFails(dEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to set up the temporary directory: %w", err)
	}

	defer tempfiles.MovableTempFileProvider.Clean()

	dataDirFS, err := dEnv.FS.WithWorkingDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to set the data directory: %w", err)
	}

	mrEnv, err := env.MultiEnvForDirectory(ctx, dEnv.Config.WriteableConfig(), dataDirFS, dEnv.Version, dEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to load database names: %w", err)
	}
	_ = mrEnv.Iter(func(dbName string, dEnv *env.DoltEnv) (stop bool, err error) {
		dsess.DefineSystemVariablesForDB(dbName)
		return false, nil
	})

	err = dsess.InitPersistedSystemVars(dEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to load persisted system variables: %w", err)
	}

	// validate that --user and --password are set appropriately.
	aprAlt, creds, err := cli.BuildUserPasswordPrompt(apr)
	apr = aprAlt
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	lateBind, err := buildLateBinder(ctx, dEnv.FS, dEnv, mrEnv, creds, apr, subcommand, false)
	if err != nil {
		return nil, err
	}

	return cli.NewCliContext(apr, dEnv.Config, lateBind)
}

// runServer launches a server on the env given and waits for it to finish
func runServer(ctx context.Context, dEnv *env.DoltEnv, args []string) error {
	// Emit a usage event in the background while we start the server.
	// Dolt is more permissive with events: it emits events even if the command fails in the earliest possible phase,
	// we emit an event only if we got far enough to attempt to launch a server (and we may not emit it if the server
	// dies quickly enough).
	//
	// We also emit a heartbeat event every 24 hours the server is running.
	// All events will be tagged with the doltgresql app id.
	go emitUsageEvent(ctx, dEnv)

	controller, err := server.RunOnDisk(ctx, args, dEnv)
	if err != nil {
		return err
	}

	return controller.WaitForStop()
}

// parseGlobalArgsAndSubCommandName parses the global arguments, including a profile if given or a default profile if exists. Also returns the subcommand name.
func parseGlobalArgsAndSubCommandName(globalConfig config.ReadWriteConfig, args []string) (apr *argparser.ArgParseResults, remaining []string, subcommandName string, err error) {
	apr, remaining, err = globalArgParser.ParseGlobalArgs(args)
	if err != nil {
		return nil, nil, "", err
	}

	subcommandName = remaining[0]

	useDefaultProfile := false
	profileName, hasProfile := apr.GetValue(commands.ProfileFlag)
	encodedProfiles, err := globalConfig.GetString(commands.GlobalCfgProfileKey)
	if err != nil {
		if err == config.ErrConfigParamNotFound {
			if hasProfile {
				return nil, nil, "", fmt.Errorf("no profiles found")
			} else {
				return apr, remaining, subcommandName, nil
			}
		} else {
			return nil, nil, "", err
		}
	}
	profiles, err := commands.DecodeProfile(encodedProfiles)
	if err != nil {
		return nil, nil, "", err
	}

	if !hasProfile {
		defaultProfile := gjson.Get(profiles, commands.DefaultProfileName)
		if defaultProfile.Exists() {
			args = append([]string{"--profile", commands.DefaultProfileName}, args...)
			apr, remaining, err = globalArgParser.ParseGlobalArgs(args)
			if err != nil {
				return nil, nil, "", err
			}
			profileName, _ = apr.GetValue(commands.ProfileFlag)
			useDefaultProfile = true
		}
	}

	if hasProfile || useDefaultProfile {
		profileArgs, err := getProfile(apr, profileName, profiles)
		if err != nil {
			return nil, nil, "", err
		}
		args = append(profileArgs, args...)
		apr, remaining, err = globalArgParser.ParseGlobalArgs(args)
		if err != nil {
			return nil, nil, "", err
		}
	}

	return
}

// getProfile retrieves the given profile from the provided list of profiles and returns the args (as flags) and values
// for that profile in a []string. If the profile is not found, an error is returned.
func getProfile(apr *argparser.ArgParseResults, profileName, profiles string) (result []string, err error) {
	prof := gjson.Get(profiles, profileName)
	if prof.Exists() {
		hasPassword := false
		password := ""
		for flag, value := range prof.Map() {
			if !apr.Contains(flag) {
				if flag == cli.PasswordFlag {
					password = value.Str
				} else if flag == "has-password" {
					hasPassword = value.Bool()
				} else if flag == cli.NoTLSFlag {
					if value.Bool() {
						result = append(result, "--"+flag)
						continue
					}
				} else {
					if value.Str != "" {
						result = append(result, "--"+flag, value.Str)
					}
				}
			}
		}
		if !apr.Contains(cli.PasswordFlag) && hasPassword {
			result = append(result, "--"+cli.PasswordFlag, password)
		}
		return result, nil
	} else {
		return nil, fmt.Errorf("profile %s not found", profileName)
	}
}

// buildLateBinder builds a LateBindQueryist for which is used to obtain the Queryist used for the length of the
// command execution.
func buildLateBinder(ctx context.Context, cwdFS filesys.Filesys, rootEnv *env.DoltEnv, mrEnv *env.MultiRepoEnv, creds *cli.UserPassword, apr *argparser.ArgParseResults, subcommandName string, verbose bool) (cli.LateBindQueryist, error) {

	var targetEnv *env.DoltEnv = nil

	useDb, hasUseDb := apr.GetValue(commands.UseDbFlag)
	useBranch, hasBranch := apr.GetValue(cli.BranchParam)

	if hasUseDb && hasBranch {
		dbName, branchNameInDb := dsess.SplitRevisionDbName(useDb)
		if len(branchNameInDb) != 0 {
			return nil, fmt.Errorf("Ambiguous branch name: %s or %s", branchNameInDb, useBranch)
		}
		useDb = dbName + "/" + useBranch
	}
	// If the host flag is given, we are forced to use a remote connection to a server.
	host, hasHost := apr.GetValue(cli.HostFlag)
	if hasHost {
		if !hasUseDb && subcommandName != "sql" {
			return nil, fmt.Errorf("The --%s flag requires the additional --%s flag.", cli.HostFlag, commands.UseDbFlag)
		}

		port, hasPort := apr.GetInt(cli.PortFlag)
		if !hasPort {
			port = 3306
		}
		useTLS := !apr.Contains(cli.NoTLSFlag)
		return sqlserver.BuildConnectionStringQueryist(ctx, cwdFS, creds, apr, host, port, useTLS, useDb)
	} else {
		_, hasPort := apr.GetInt(cli.PortFlag)
		if hasPort {
			return nil, fmt.Errorf("The --%s flag is only meaningful with the --%s flag.", cli.PortFlag, cli.HostFlag)
		}
	}

	if hasUseDb {
		dbName, _ := dsess.SplitRevisionDbName(useDb)
		targetEnv = mrEnv.GetEnv(dbName)
		if targetEnv == nil {
			return nil, fmt.Errorf("The provided --use-db %s does not exist.", dbName)
		}
	} else {
		useDb = mrEnv.GetFirstDatabase()
		if hasBranch {
			useDb += "/" + useBranch
		}
	}

	if targetEnv == nil && useDb != "" {
		targetEnv = mrEnv.GetEnv(useDb)
	}

	// There is no target environment detected. This is allowed for a small number of commands.
	// We don't expect that number to grow, so we list them here.
	// It's also allowed when --help is passed.
	// So we defer the error until the caller tries to use the cli.LateBindQueryist
	isDoltEnvironmentRequired := subcommandName != "init" && subcommandName != "sql" && subcommandName != "sql-server" && subcommandName != "sql-client"
	if targetEnv == nil && isDoltEnvironmentRequired {
		return func(ctx context.Context) (cli.Queryist, *sql.Context, func(), error) {
			return nil, nil, nil, fmt.Errorf("The current directory is not a valid dolt repository.")
		}, nil
	}

	if verbose {
		cli.Println("verbose: starting local mode")
	}
	return commands.BuildSqlEngineQueryist(ctx, cwdFS, mrEnv, creds, apr)
}

// If we cannot verify that we can move files for any reason, use a ./.dolt/tmp as the temp dir.
func reconfigIfTempFileMoveFails(dEnv *env.DoltEnv) error {
	if !canMoveTempFile() {
		tmpDir := "./.dolt/tmp"

		if !dEnv.HasDoltDir() {
			tmpDir = "./.tmp"
		}

		stat, err := os.Stat(tmpDir)

		if err != nil {
			err := os.MkdirAll(tmpDir, os.ModePerm)

			if err != nil {
				return fmt.Errorf("failed to create temp dir '%s': %s", tmpDir, err.Error())
			}
		} else if !stat.IsDir() {
			return fmt.Errorf("attempting to use '%s' as a temp directory, but there exists a file with that name", tmpDir)
		}

		tempfiles.MovableTempFileProvider = tempfiles.NewTempFileProviderAt(tmpDir)
	}

	return nil
}

func canMoveTempFile() bool {
	const testfile = "./testfile"

	f, err := os.CreateTemp("", "")

	if err != nil {
		return false
	}

	name := f.Name()
	err = f.Close()

	if err != nil {
		return false
	}

	err = file.Rename(name, testfile)

	if err != nil {
		_ = file.Remove(name)
		return false
	}

	_ = file.Remove(testfile)
	return true
}

func redirectStdio(args []string) ([]string, error) {
	if len(args) > 0 {
		var doneDebugFlags bool
		for !doneDebugFlags && len(args) > 0 {
			switch args[0] {
			// Currently goland doesn't support running with a different working directory when using go modules.
			// This is a hack that allows a different working directory to be set after the application starts using
			// chdir=<DIR>.  The syntax is not flexible and must match exactly this.
			case chdirFlag:
				err := os.Chdir(args[1])

				if err != nil {
					panic(err)
				}

				args = args[2:]

			case stdInFlag:
				stdInFile := args[1]
				cli.Println("Using file contents as stdin:", stdInFile)

				f, err := os.Open(stdInFile)
				if err != nil {
					return nil, fmt.Errorf("Failed to open %s: %w", stdInFile, err)
				}

				os.Stdin = f
				args = args[2:]

			case stdOutFlag, stdErrFlag, stdOutAndErrFlag:
				filename := args[1]

				f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
				if err != nil {
					return nil, fmt.Errorf("Failed to open %s for writing: %w", filename, err)
				}

				switch args[0] {
				case stdOutFlag:
					cli.Println("Stdout being written to", filename)
					cli.CliOut = f
				case stdErrFlag:
					cli.Println("Stderr being written to", filename)
					cli.CliErr = f
				case stdOutAndErrFlag:
					cli.Println("Stdout and Stderr being written to", filename)
					cli.CliOut = f
					cli.CliErr = f
				}

				color.NoColor = true
				args = args[2:]

			default:
				doneDebugFlags = true
			}
		}
	}
	return args, nil
}

// emitUsageEvent emits a usage event to the event server
func emitUsageEvent(ctx context.Context, dEnv *env.DoltEnv) {
	metricsDisabled := dEnv.Config.GetStringOrDefault(config.MetricsDisabled, "false")
	disabled, err := strconv.ParseBool(metricsDisabled)
	if err != nil || disabled {
		return
	}

	emitter, err := commands.GRPCEmitterForConfig(dEnv)
	if err != nil {
		return
	}

	evt := events.NewEvent(sqlserver.SqlServerCmd{}.EventType())
	evtCollector := events.NewCollector(server.Version, emitter)
	evtCollector.CloseEventAndAdd(evt)
	clientEvents := evtCollector.Close()

	_ = emitter.LogEvents(ctx, server.Version, clientEvents)
}
