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
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"

	"github.com/dolthub/dolt/go/cmd/dolt/cli"
	"github.com/dolthub/dolt/go/cmd/dolt/commands"
	"github.com/dolthub/dolt/go/cmd/dolt/commands/sqlserver"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/utils/argparser"
	"github.com/dolthub/dolt/go/libraries/utils/config"
	"github.com/dolthub/dolt/go/libraries/utils/file"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/dolt/go/store/util/tempfiles"
	"github.com/dolthub/doltgresql/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/tidwall/gjson"
)


var doltgresCommands = cli.NewSubCommandHandler("doltgresql", "it's git for data", []cli.Command{
	commands.ConfigCmd{},
	commands.VersionCmd{VersionStr: server.Version},
	sqlserver.SqlServerCmd{VersionStr: server.Version},
})
var globalArgParser = cli.CreateGlobalArgParser("doltgresql")

func main() {
	ctx := context.Background()
	seedGlobalRand()

	restoreIO := cli.InitIO()
	defer restoreIO()

	warnIfMaxFilesTooLow()

	fs := filesys.LocalFS
	dEnv := env.Load(ctx, env.GetCurrentUserHomeDir, fs, doltdb.LocalDirDoltDB, server.Version)

	globalConfig, _ := dEnv.Config.GetConfig(env.GlobalConfig)
	apr, args, subCommandName, err := parseGlobalArgsAndSubCommandName(globalConfig, os.Args[1:])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	cliCtx, err := configureCliCtx(apr, fs, dEnv, err, ctx)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// the sql-server command has special cased logic since we have to wait for the server to stop
	if subCommandName == "" || subCommandName == "sql-server" {
		runServer(ctx, dEnv, cliCtx)
	}

	exitCode := doltgresCommands.Exec(ctx, "doltgresql", args, dEnv, cliCtx)
	os.Exit(exitCode)
}

func configureCliCtx(apr *argparser.ArgParseResults, fs filesys.Filesys, dEnv *env.DoltEnv, err error, ctx context.Context) (cli.CliContext, error) {
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
		return nil, fmt.Errorf("Cannot start a server within a directory containing a Dolt or Doltgres database." +
				"To use the current directory as a database, start the server from the parent directory.")
	}

	err = reconfigIfTempFileMoveFails(dEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to set up the temporary directory: %w", err)
	}

	defer tempfiles.MovableTempFileProvider.Clean()

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

	return cli.NewCliContext(apr, dEnv.Config, lateBind)
}

func runServer(ctx context.Context, dEnv *env.DoltEnv, cliCtx cli.CliContext) {
	controller, err := server.RunOnDisk(ctx, os.Args[1:], dEnv, cliCtx)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = controller.WaitForStop()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
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

func seedGlobalRand() {
	bs := make([]byte, 8)
	_, err := crand.Read(bs)
	if err != nil {
		panic("failed to initial rand " + err.Error())
	}
	rand.Seed(int64(binary.LittleEndian.Uint64(bs)))
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
