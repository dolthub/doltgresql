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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/cmd/dolt/cli"
	"github.com/dolthub/dolt/go/cmd/dolt/commands/sqlserver"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	doltservercfg "github.com/dolthub/dolt/go/libraries/doltcore/servercfg"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dfunctions"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/dolt/go/libraries/utils/argparser"
	"github.com/dolthub/dolt/go/libraries/utils/config"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/dolthub/dolt/go/store/util/tempfiles"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"

	"github.com/dolthub/doltgresql/server/initialization"
	"github.com/dolthub/doltgresql/server/logrepl"
	"github.com/dolthub/doltgresql/servercfg"
)

// Version should have a new line that follows, else the formatter will fail the PR created by the release GH action

const (
	Version = "0.51.2"

	DefUserName  = "postres"
	DefUserEmail = "postgres@somewhere.com"
	DoltgresDir  = "postgres"
)

func init() {
	sqlserver.ExternalDisableUsers = true
	dfunctions.VersionString = Version
	resolve.UseSearchPath = true
}

// RunOnDisk starts the server based on the given args, while also using the local disk as the backing store.
// The returned WaitGroup may be used to wait for the server to close.
func RunOnDisk(ctx context.Context, cfg *servercfg.DoltgresConfig, dEnv *env.DoltEnv) (*svcs.Controller, error) {
	return runServer(ctx, cfg, dEnv, NewListener)
}

// RunInMemory starts the server based on the given args, while also using RAM as the backing store.
// The returned WaitGroup may be used to wait for the server to close.
func RunInMemory(cfg *servercfg.DoltgresConfig, protocolListenerFactory server.ProtocolListenerFunc) (*svcs.Controller, error) {
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

	return runServer(ctx, cfg, dEnv, protocolListenerFactory)
}

// runServer starts the server based on the given args, using the provided file system as the backing store.
// The returned WaitGroup may be used to wait for the server to close.
func runServer(ctx context.Context, cfg *servercfg.DoltgresConfig, dEnv *env.DoltEnv, protocolListenerFactory server.ProtocolListenerFunc) (*svcs.Controller, error) {
	initialization.Initialize(dEnv)

	if dEnv.HasDoltDataDir() {
		cwd, _ := dEnv.FS.Abs(".")
		return nil, errors.Errorf("Cannot start a server within a directory containing a Dolt or Doltgres database. "+
			"To use the current directory (%s) as a database, start the server from the parent directory.", cwd)
	}

	defer tempfiles.MovableTempFileProvider.Clean()

	sql.SystemVariables.SetGlobal(sql.NewContext(ctx), dsess.DoltStatsEnabled, false)

	err := dsess.InitPersistedSystemVars(dEnv)
	if err != nil {
		return nil, errors.Errorf("failed to load persisted system variables: %w", err)
	}

	ssCfg := cfg.ToSqlServerConfig()
	// The sql context can't be passed in because doesn't exist yet.
	// But since it's only needed to read from the db and the db doesn't exist yet either, this is safe.
	err = doltservercfg.ApplySystemVariables(nil, ssCfg, sql.SystemVariables)
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

	// Reload the dolt environment with the correct data dir that was specified in the configuration.
	// This initial dEnv instance is loaded for the current working directory.
	dataDirFs, err := dEnv.FS.WithWorkingDir(ssCfg.DataDir())
	if err != nil {
		return nil, err
	}
	dEnv = env.Load(ctx, dEnv.GetUserHomeDir, dataDirFs, doltdb.LocalDirDoltDB, dEnv.Version)

	// Automatically initialize a doltgres database if necessary
	// TODO: probably should only do this if there are no databases in the data dir already
	createDoltgresDatabase := false
	if exists, isDirectory := dataDirFs.Exists(DoltgresDir); !exists {
		createDoltgresDatabase = true
	} else if !isDirectory {
		workingDir, _ := dataDirFs.Abs(".")
		// The else branch means that there's a Doltgres item, so we need to error if it's a file since we
		// enforce the creation of a Doltgres database/directory, which would create a name conflict with the file
		return nil, errors.Errorf("Attempted to create the default `postgres` database at `%s`, but a file with "+
			"the same name was found. Either remove the file, change the directory using the `--data-dir` argument, "+
			"or change the environment variable `%s` so that it points to a different directory.", workingDir, servercfg.DOLTGRES_DATA_DIR)
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

	sqlserver.ConfigureServices(&sqlserver.Config{
		Version:                 Version,
		ServerConfig:            ssCfg,
		Controller:              controller,
		DoltEnv:                 dEnv,
		ProtocolListenerFactory: protocolListenerFactory,
	})
	go controller.Start(newCtx)

	err = controller.WaitForStart()
	if err != nil {
		return nil, err
	}

	if createDoltgresDatabase {
		err = createDatabase(ssCfg, "postgres")
		if err != nil {
			return nil, err
		}
	}

	// TODO: shutdown replication cleanly when we stop the server
	_, err = startReplication(cfg, ssCfg)
	if err != nil {
		return nil, err
	}

	return controller, nil
}

// createDatabase creates the database named on the local server using the configuration values to connect, returning
// any error
func createDatabase(cfg doltservercfg.ServerConfig, dbName string) error {
	dsn := fmt.Sprintf("postgres://postgres:password@localhost:%d", cfg.Port())

	// Connect to the server and create the default database with the given name.
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", dbName))
	return err
}

// startReplication begins the background thread that replicates from Postgres, if one is configured.
func startReplication(cfg *servercfg.DoltgresConfig, ssCfg doltservercfg.ServerConfig) (*logrepl.LogicalReplicator, error) {
	if cfg.PostgresReplicationConfig == nil {
		return nil, nil
	} else if cfg.PostgresReplicationConfig.PostgresDatabase == nil || *cfg.PostgresReplicationConfig.PostgresDatabase == "" {
		return nil, errors.Errorf("postgres replication database must be specified and not empty for replication")
	} else if cfg.PostgresReplicationConfig.PostgresUser == nil || *cfg.PostgresReplicationConfig.PostgresUser == "" {
		return nil, errors.Errorf("postgres replication user must be specified and not empty for replication")
	} else if cfg.PostgresReplicationConfig.PostgresPassword == nil || *cfg.PostgresReplicationConfig.PostgresPassword == "" {
		return nil, errors.Errorf("postgres replication password must be specified and not empty for replication")
	} else if cfg.PostgresReplicationConfig.PostgresPort == nil || *cfg.PostgresReplicationConfig.PostgresPort == 0 {
		return nil, errors.Errorf("postgres replication port must be specified and non-zero for replication")
	} else if cfg.PostgresReplicationConfig.SlotName == nil || *cfg.PostgresReplicationConfig.SlotName == "" {
		return nil, errors.Errorf("postgres replication slot name must be specified and not empty for replication")
	}

	walFilePath := filepath.Join(ssCfg.CfgDir(), "pg_wal_location")
	primaryDns := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		*cfg.PostgresReplicationConfig.PostgresUser,
		*cfg.PostgresReplicationConfig.PostgresPassword,
		*cfg.PostgresReplicationConfig.PostgresServerAddress,
		*cfg.PostgresReplicationConfig.PostgresPort,
		*cfg.PostgresReplicationConfig.PostgresDatabase,
	)

	replicationDns := fmt.Sprintf(
		"postgres://%s:%s@localhost:%d/%s",
		ssCfg.User(),
		ssCfg.Password(),
		ssCfg.Port(),
		"postgres", // TODO: this needs to come from config
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

func (c configCliContext) QueryEngine(ctx context.Context) (cli.QueryEngineResult, error) {
	return cli.QueryEngineResult{}, errors.Errorf("ConfigCliContext does not support QueryEngine()")
}

func (c configCliContext) WorkingDir() filesys.Filesys {
	panic("runtime error:ConfigCliContext does not support WorkingDir() in this context")
}

func (c configCliContext) Close() {}

var _ cli.CliContext = configCliContext{}
