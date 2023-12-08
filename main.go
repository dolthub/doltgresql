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
	"github.com/dolthub/dolt/go/libraries/utils/argparser"
	"github.com/dolthub/dolt/go/libraries/utils/config"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/doltgresql/server"
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

	if subCommandName == "" || subCommandName == "sql-server" {
		runServer(ctx, dEnv)
	}

	lateBind, err := buildLateBinder(ctx, cwdFS, dEnv, mrEnv, creds, apr, "sql-server", false)
	if err != nil {
		return nil, err
	}

	cliCtx, err := cli.NewCliContext(apr, dEnv.Config, lateBind)
	if err != nil {
		return nil, err
	}

	doltgresCommands.Exec(ctx, "doltgresql", args, dEnv)

	os.Exit(0)
}

func runServer(ctx context.Context, dEnv *env.DoltEnv) {
	controller, err := server.RunOnDisk(ctx, os.Args[1:], dEnv)

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