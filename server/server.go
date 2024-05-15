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
	"fmt"
	_ "net/http/pprof"
	"path/filepath"

	"github.com/dolthub/dolt/go/cmd/dolt/cli"
	"github.com/dolthub/dolt/go/cmd/dolt/commands"
	"github.com/dolthub/dolt/go/cmd/dolt/commands/sqlserver"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	doltservercfg "github.com/dolthub/dolt/go/libraries/doltcore/servercfg"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dfunctions"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/utils/argparser"
	"github.com/dolthub/dolt/go/libraries/utils/config"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/dolthub/dolt/go/store/util/tempfiles"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/initialization"
	"github.com/dolthub/doltgresql/server/logrepl"
	"github.com/dolthub/doltgresql/servercfg"
)

const (
	Version = "0.7.4"

	// DOLTGRES_DATA_DIR is an environment variable that defines the location of DoltgreSQL databases
	DOLTGRES_DATA_DIR = "DOLTGRES_DATA_DIR"
	// DOLTGRES_DATA_DIR_DEFAULT is the portion to append to the user's home directory if DOLTGRES_DATA_DIR has not been specified
	DOLTGRES_DATA_DIR_DEFAULT = "doltgres/databases"

	DefUserName  = "postres"
	DefUserEmail = "postgres@somewhere.com"
	DoltgresDir  = "doltgres"
)

func init() {
	server.DefaultProtocolListenerFunc = NewListener
	sqlserver.ExternalDisableUsers = true
	dfunctions.VersionString = Version
}

// RunOnDisk starts the server based on the given args, while also using the local disk as the backing store.
// The returned WaitGroup may be used to wait for the server to close.
func RunOnDisk(ctx context.Context, cfg *servercfg.DoltgresConfig, dEnv *env.DoltEnv) (*svcs.Controller, error) {
	return runServer(ctx, cfg, dEnv)
}

// RunInMemory starts the server based on the given args, while also using RAM as the backing store.
// The returned WaitGroup may be used to wait for the server to close.
func RunInMemory(cfg *servercfg.DoltgresConfig) (*svcs.Controller, error) {
	ctx := context.Background()
	fs := filesys.EmptyInMemFS("")
	dEnv := env.Load(ctx, env.GetCurrentUserHomeDir, fs, doltdb.InMemDoltDB, Version)
	globalConfig, _ := dEnv.Config.GetConfig(env.GlobalConfig)
	if globalConfig.GetStringOrDefault(config.UserNameKey, "") == "" {
		globalConfig.SetStrings(map[string]string{
			config.UserNameKey:  DefUserName,
			config.UserEmailKey: DefUserEmail,
		})
	}

	return runServer(ctx, cfg, dEnv)
}

// runServer starts the server based on the given args, using the provided file system as the backing store.
// The returned WaitGroup may be used to wait for the server to close.
func runServer(ctx context.Context, cfg *servercfg.DoltgresConfig, dEnv *env.DoltEnv) (*svcs.Controller, error) {
	initialization.Initialize()

	if dEnv.HasDoltDataDir() {
		cwd, _ := dEnv.FS.Abs(".")
		return nil, fmt.Errorf("Cannot start a server within a directory containing a Dolt or Doltgres database. "+
			"To use the current directory (%s) as a database, start the server from the parent directory.", cwd)
	}

	defer tempfiles.MovableTempFileProvider.Clean()

	err := dsess.InitPersistedSystemVars(dEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to load persisted system variables: %w", err)
	}

	ssCfg := cfg.ToSqlServerConfig()
	err = doltservercfg.ApplySystemVariables(ssCfg, sql.SystemVariables)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := doltservercfg.LoadTLSConfig(ssCfg)
	if err != nil {
		return nil, err
	}

	if tlsConfig != nil && len(tlsConfig.Certificates) > 0 {
		certificate = tlsConfig.Certificates[0]
	}

	// We need a username and password for many SQL commands, so set defaults if they don't exist
	dEnv.Config.SetFailsafes(map[string]string{
		config.UserNameKey:  DefUserName,
		config.UserEmailKey: DefUserEmail,
	})

	dataDirFs, err := filesys.LocalFS.WithWorkingDir(ssCfg.DataDir())
	if err != nil {
		return nil, err
	}

	// Automatically initialize a doltgres database if necessary
	// TODO: probably should only do this if there are no databases in the data dir already
	if exists, isDirectory := dataDirFs.Exists(DoltgresDir); !exists {
		err := dataDirFs.MkDirs(DoltgresDir)
		if err != nil {
			return nil, err
		}
		subdirectoryFS, err := dataDirFs.WithWorkingDir(DoltgresDir)
		if err != nil {
			return nil, err
		}

		// We'll use a temporary environment to instantiate the subdirectory
		tempDEnv := env.Load(ctx, env.GetCurrentUserHomeDir, subdirectoryFS, dEnv.UrlStr(), Version)
		// username and user email is needed to create a new database.
		name := tempDEnv.Config.GetStringOrDefault(config.UserNameKey, DefUserName)
		email := tempDEnv.Config.GetStringOrDefault(config.UserEmailKey, DefUserEmail)
		res := commands.InitCmd{}.Exec(ctx, "init", []string{"--name", name, "--email", email}, tempDEnv, configCliContext{tempDEnv})
		if res != 0 {
			return nil, fmt.Errorf("failed to initialize doltgres database")
		}
	} else if !isDirectory {
		workingDir, _ := dataDirFs.Abs(".")
		// The else branch means that there's a Doltgres item, so we need to error if it's a file since we
		// enforce the creation of a Doltgres database/directory, which would create a name conflict with the file
		return nil, fmt.Errorf("Attempted to create the default `doltgres` database at `%s`, but a file with "+
			"the same name was found. Either remove the file, change the directory using the `--data-dir` argument, "+
			"or change the environment variable `%s` so that it points to a different directory.", workingDir, DOLTGRES_DATA_DIR)
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

	sqlserver.ConfigureServices(ssCfg, controller, Version, dEnv)
	go controller.Start(newCtx)

	err = controller.WaitForStart()
	if err != nil {
		return nil, err
	}

	// TODO: shutdown replication cleanly when we stop the server
	_, err = startReplication(cfg, ssCfg)
	if err != nil {
		return nil, err
	}

	return controller, nil
}

// startReplication begins the background thread that replicates from Postgres, if one is configured.
func startReplication(cfg *servercfg.DoltgresConfig, ssCfg doltservercfg.ServerConfig) (*logrepl.LogicalReplicator, error) {
	if cfg.PostgresReplicationConfig == nil {
		return nil, nil
	} else if cfg.PostgresReplicationConfig.PostgresDatabase == nil || *cfg.PostgresReplicationConfig.PostgresDatabase == "" {
		return nil, fmt.Errorf("postgres replication database must be specified and not empty for replication")
	} else if cfg.PostgresReplicationConfig.PostgresUser == nil || *cfg.PostgresReplicationConfig.PostgresUser == "" {
		return nil, fmt.Errorf("postgres replication user must be specified and not empty for replication")
	} else if cfg.PostgresReplicationConfig.PostgresPassword == nil || *cfg.PostgresReplicationConfig.PostgresPassword == "" {
		return nil, fmt.Errorf("postgres replication password must be specified and not empty for replication")
	} else if cfg.PostgresReplicationConfig.PostgresPort == nil || *cfg.PostgresReplicationConfig.PostgresPort == 0 {
		return nil, fmt.Errorf("postgres replication port must be specified and non-zero for replication")
	} else if cfg.PostgresReplicationConfig.SlotName == nil || *cfg.PostgresReplicationConfig.SlotName == "" {
		return nil, fmt.Errorf("postgres replication slot name must be specified and not empty for replication")
	}

	walFilePath := filepath.Join(ssCfg.CfgDir(), "pg_wal_location")
	primaryDns := fmt.Sprintf(
		"postgres://%s:%s@127.0.0.1:%d/%s",
		*cfg.PostgresReplicationConfig.PostgresUser,
		*cfg.PostgresReplicationConfig.PostgresPassword,
		*cfg.PostgresReplicationConfig.PostgresPort,
		*cfg.PostgresReplicationConfig.PostgresDatabase,
	)

	replicationDns := fmt.Sprintf(
		"postgres://%s:%s@localhost:%d/%s",
		ssCfg.User(),
		ssCfg.Password(),
		ssCfg.Port(),
		"doltgres", // TODO: this needs to come from config
	)

	replicator, err := logrepl.NewLogicalReplicator(walFilePath, primaryDns, replicationDns)
	if err != nil {
		return nil, err
	}

	cli.Println("Starting replication")
	go replicator.StartReplication(*cfg.PostgresReplicationConfig.SlotName)
	return replicator, nil
}

// configCliContext is a minimal implementation of CliContext that only supports Config()
type configCliContext struct {
	dEnv *env.DoltEnv
}

func (c configCliContext) Config() *env.DoltCliConfig {
	return c.dEnv.Config
}

func (c configCliContext) GlobalArgs() *argparser.ArgParseResults {
	panic("ConfigCliContext does not support GlobalArgs()")
}

func (c configCliContext) QueryEngine(ctx context.Context) (cli.Queryist, *sql.Context, func(), error) {
	return nil, nil, nil, fmt.Errorf("ConfigCliContext does not support QueryEngine()")
}

var _ cli.CliContext = configCliContext{}
