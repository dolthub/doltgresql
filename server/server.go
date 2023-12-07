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

package server

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	_ "net/http/pprof"
	"os"

	"github.com/dolthub/dolt/go/cmd/dolt/cli"
	"github.com/dolthub/dolt/go/cmd/dolt/commands"
	"github.com/dolthub/dolt/go/cmd/dolt/commands/sqlserver"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dfunctions"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/utils/argparser"
	"github.com/dolthub/dolt/go/libraries/utils/config"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/dolthub/dolt/go/store/nbs"
	"github.com/dolthub/dolt/go/store/util/tempfiles"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/fatih/color"
	"github.com/tidwall/gjson"
)

//TODO: cleanup this file

const (
	Version = "0.1.0"
)

var doltCommand = cli.NewSubCommandHandler("doltgresql", "it's git for data", []cli.Command{
	commands.InitCmd{},
	commands.ConfigCmd{},
	commands.VersionCmd{VersionStr: Version},
	sqlserver.SqlServerCmd{VersionStr: Version},
})
var globalArgParser = cli.CreateGlobalArgParser("doltgresql")

func init() {
	server.DefaultProtocolListenerFunc = NewListener
	sqlserver.ExternalDisableUsers = true
	dfunctions.VersionString = Version
}

const chdirFlag = "--chdir"
const stdInFlag = "--stdin"
const stdOutFlag = "--stdout"
const stdErrFlag = "--stderr"
const stdOutAndErrFlag = "--out-and-err"

// RunOnDisk starts the server based on the given args, while also using the local disk as the backing store.
// The returned WaitGroup may be used to wait for the server to close.
func RunOnDisk(args []string) (*svcs.Controller, error) {
	return runServer(args, filesys.LocalFS)
}

// RunInMemory starts the server based on the given args, while also using RAM as the backing store.
// The returned WaitGroup may be used to wait for the server to close.
func RunInMemory(args []string) (*svcs.Controller, error) {
	return runServer(args, filesys.EmptyInMemFS(""))
}

// runServer starts the server based on the given args, using the provided file system as the backing store.
// The returned WaitGroup may be used to wait for the server to close.
func runServer(args []string, fs filesys.Filesys) (*svcs.Controller, error) {
	ctx := context.Background()

	if serverArgs, err := (sqlserver.SqlServerCmd{}).ArgParser().Parse(append([]string{"sql-server"}, args...)); err == nil {
		if _, ok := serverArgs.GetValue("port"); !ok {
			args = append(args, "--port=5432")
		}
	}

	if os.Getenv("DOLT_VERBOSE_ASSERT_TABLE_FILES_CLOSED") == "" {
		nbs.TableIndexGCFinalizerWithStackTrace = false
	}

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

	seedGlobalRand()

	restoreIO := cli.InitIO()
	defer restoreIO()

	warnIfMaxFilesTooLow()

	dEnv := env.Load(ctx, env.GetCurrentUserHomeDir, fs, doltdb.LocalDirDoltDB, Version)

	globalConfig, ok := dEnv.Config.GetConfig(env.GlobalConfig)
	if !ok {
		return nil, fmt.Errorf("Failed to get global config")
	}
	// The in-memory database is only used for testing/virtual environments, so the config may need modification
	if _, ok = fs.(*filesys.InMemFS); ok && globalConfig.GetStringOrDefault(config.UserNameKey, "") == "" {
		globalConfig.SetStrings(map[string]string{
			config.UserNameKey:  "postgres",
			config.UserEmailKey: "postgres@somewhere.com",
		})
	}

	apr, _, _, err := parseGlobalArgsAndSubCommandName(globalConfig, append([]string {"sql-server"}, args...))
	if err == argparser.ErrHelp {
		//TODO: display some help message
		return nil, fmt.Errorf("help")
	} else if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	dataDir, hasDataDir := apr.GetValue(commands.DataDirFlag)
	if hasDataDir {
		// If a relative path was provided, this ensures we have an absolute path everywhere.
		dataDir, err = fs.Abs(dataDir)
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
		return nil, fmt.Errorf("Cannot start a server within a directory containing a Dolt or Doltgres database." +
				"To use the current directory as a database, start the server from the parent directory.")
	}

	err = reconfigIfTempFileMoveFails(dEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to set up the temporary directory: %w", err)
	}

	defer tempfiles.MovableTempFileProvider.Clean()

	// Find all database names and add global variables for them. This needs to
	// occur before a call to dsess.InitPersistedSystemVars. Otherwise, database
	// specific persisted system vars will fail to load.
	//
	// In general, there is a lot of work TODO in this area. System global
	// variables are persisted to the Dolt local config if found and if not
	// found the Dolt global config (typically ~/.dolt/config_global.json).

	// Depending on what directory a dolt sql-server is started in, users may
	// see different variables values. For example, start a dolt sql-server in
	// the dolt database folder and persist some system variable.

	// If dolt sql-server is started outside that folder, those system variables
	// will be lost. This is particularly confusing for database specific system
	// variables like `${db_name}_default_branch` (maybe these should not be
	// part of Dolt config in the first place!).

	// Current working directory is preserved to ensure that user provided path arguments are always calculated
	// relative to this directory. The root environment's FS will be updated to be the --data-dir path if the user
	// specified one.
	cwdFS := dEnv.FS
	dataDirFS, err := dEnv.FS.WithWorkingDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to set the data directory: %w", err)
	}
	dEnv.FS = dataDirFS

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

	lateBind, err := buildLateBinder(ctx, cwdFS, dEnv, mrEnv, creds, apr, "sql-server", false)
	if err != nil {
		return nil, err
	}

	cliCtx, err := cli.NewCliContext(apr, dEnv.Config, lateBind)
	if err != nil {
		return nil, err
	}
	
	serverConfig, err := sqlserver.ServerConfigFromArgs("sql-server", args, dEnv)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := sqlserver.LoadTLSConfig(serverConfig)
	if err != nil {
		return nil, err
	}
	
	if tlsConfig != nil && len(tlsConfig.Certificates) > 0 {
		certificate = tlsConfig.Certificates[0]
	}

	// Automatically initializes a doltgres database if necessary
	if !dEnv.HasDoltDir() {
		// Need to make sure that there isn't a doltgres subdirectory. If there is, we'll assume it's a db.
		if exists, _ := dEnv.FS.Exists("doltgres"); !exists {
			err := dEnv.FS.MkDirs("doltgres")
			if err != nil {
				return nil, err
			}
			subdirectoryFS, err := dEnv.FS.WithWorkingDir("doltgres")
			if err != nil {
				return nil, err
			}
			// We'll use a temporary environment to instantiate the subdirectory
			tempDEnv := env.Load(ctx, env.GetCurrentUserHomeDir, subdirectoryFS, doltdb.LocalDirDoltDB, Version)
			_ = doltCommand.Exec(ctx, "dolt", []string{"init"}, tempDEnv, cliCtx)
		}
	}

	controller := svcs.NewController()
	newCtx, cancelF := context.WithCancel(ctx)
	go func() {
		// Here we only forward along the SIGINT if the server starts
		// up successfully.  If the service does not start up
		// successfully, or if WaitForStart() blocks indefinitely, then
		// startServer() should have returned an error and we do not
		// need to Stop the running server or deal with our canceled
		// parent context.
		if controller.WaitForStart() == nil {
			<-ctx.Done()
			controller.Stop()
			cancelF()
		}
	}()

	// We need a username and password for many SQL commands, so set defaults if they don't exist
	dEnv.Config.SetFailsafes(env.DefaultFailsafeConfig)
	
	sqlserver.ConfigureServices(serverConfig, controller, Version, dEnv)
	go controller.Start(newCtx)
	return controller, controller.WaitForStart()
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

	// nil targetEnv will happen if the user ran a command in an empty directory or when there is a server running with
	// no databases. CLI will try to connect to the server in this case.
	if targetEnv == nil {
		targetEnv = rootEnv
	}
	
	if verbose {
		cli.Println("verbose: starting local mode")
	}
	return commands.BuildSqlEngineQueryist(ctx, cwdFS, mrEnv, creds, apr)
}

func seedGlobalRand() {
	bs := make([]byte, 8)
	_, err := crand.Read(bs)
	if err != nil {
		panic("failed to initial rand " + err.Error())
	}
	rand.Seed(int64(binary.LittleEndian.Uint64(bs)))
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

func intPointer(val int) *int {
	p := new(int)
	*p = val
	return p
}
