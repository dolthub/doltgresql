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
	"net"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"time"

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

	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/initialization"
	"github.com/dolthub/doltgresql/server/logrepl"
	"github.com/dolthub/doltgresql/servercfg"
)

// Version should have a new line that follows, else the formatter will fail the PR created by the release GH action

const (
	Version = "0.56.7"

	DefUserName         = "postres"
	DefUserEmail        = "postgres@somewhere.com"
	DefUserEmailFmt     = "%s@somewhere.com"
	DefaultDbNameEnvVar = "DOLTGRES_DB"
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
	initialization.Initialize(dEnv, cfg)

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
	user, _ := auth.GetSuperUserAndPassword()
	dEnv.Config.SetFailsafes(map[string]string{
		config.UserNameKey:  user,
		config.UserEmailKey: fmt.Sprintf(DefUserEmailFmt, user),
	})

	// Reload the dolt environment with the correct data dir that was specified in the configuration.
	// This initial dEnv instance is loaded for the current working directory.
	dataDirFs, err := dEnv.FS.WithWorkingDir(ssCfg.DataDir())
	if err != nil {
		return nil, err
	}
	dEnv = env.Load(ctx, dEnv.GetUserHomeDir, dataDirFs, doltdb.LocalDirDoltDB, dEnv.Version)

	// Determine whether we need to initialize the default database
	initializeDefaultDatabase := true
	mrEnv, err := env.MultiEnvForDirectory(ctx, dataDirFs, dEnv)
	if err != nil {
		return nil, err
	}
	mrEnv.Iter(func(_ string, _ *env.DoltEnv) (stop bool, err error) {
		initializeDefaultDatabase = false
		return true, nil
	})

	// When we need to create the default database, gate the listener so that no external connections are accepted until
	// that creation has fully completed. This is necessary due to the fact that `CREATE DATABASE` is non-transactional
	// and non-atomic, so other clients could connect to a partially-created database otherwise. This problem exists in
	// Dolt as well, but Dolt doesn't auto-create a database on init.
	var gate *startupGate
	if initializeDefaultDatabase {
		gate = newStartupGate()
		protocolListenerFactory = gate.gatedListenerFactory(protocolListenerFactory)
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
		ProviderFactory:         DoltgresProviderFactory{},
	})
	go controller.Start(newCtx)

	err = controller.WaitForStart()
	if err != nil {
		return nil, err
	}

	if initializeDefaultDatabase {
		err = createDefaultDatabase(ssCfg, gate)
		if err != nil {
			controller.Stop()
			return nil, err
		}
		// Release the startup gate to accept external connections now that init is finished
		gate.Release()
	}

	// TODO: shutdown replication cleanly when we stop the server
	_, err = startReplication(cfg, ssCfg)
	if err != nil {
		controller.Stop()
		return nil, err
	}

	return controller, nil
}

// createDefaultDatabaseTimeout bounds first-run creation of the default
// database. The internal connection is served by the same accept loop that
// serves clients, so if the server fails to start its accept loop the dial
// would otherwise block forever.
const createDefaultDatabaseTimeout = 2 * time.Minute

// createDefaultDatabase creates the database named on the local server, returning any error. The connection used is an
// internal in-memory connection provided by |gate|; the server does not accept external connections until the gate is
// released, which the caller does only after this function succeeds.
func createDefaultDatabase(cfg doltservercfg.ServerConfig, gate *startupGate) error {
	user, password := auth.GetSuperUserAndPassword()
	dbName := getDefaultDatabaseName(user)

	// The host and port here are only used for display purposes: DialFunc below routes the connection through the
	// gate's in-memory pipe. TLS is disabled because it is unnecessary for an in-process connection.
	dsn := fmt.Sprintf("postgres://%s:%s@localhost:%d/?sslmode=disable", user, password, cfg.Port())

	ctx, cancel := context.WithTimeout(context.Background(), createDefaultDatabaseTimeout)
	defer cancel()

	connConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return err
	}
	connConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return gate.Dial(ctx)
	}

	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", dbName))
	return err
}

// getDefaultDatabaseName returns the name of the default database to create on first server start.
// If the environment variable DOLTGRES_DB is set, that value is used. Otherwise, the username is used.
// The username is in turn configured with the environment variable DOLTGRES_USER, defaulting to "postgres".
func getDefaultDatabaseName(userName string) string {
	defaultDbName := os.Getenv(DefaultDbNameEnvVar)
	if defaultDbName != "" {
		return defaultDbName
	}
	return userName
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

func (c configCliContext) QueryEngine(ctx context.Context, _ ...cli.LateBindQueryistOption) (cli.QueryEngineResult, error) {
	return cli.QueryEngineResult{}, errors.Errorf("ConfigCliContext does not support QueryEngine()")
}

func (c configCliContext) WorkingDir() filesys.Filesys {
	panic("runtime error:ConfigCliContext does not support WorkingDir() in this context")
}

func (c configCliContext) Close() {}

var _ cli.CliContext = configCliContext{}
