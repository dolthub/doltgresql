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
	"strings"

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
	"github.com/dolthub/dolt/go/store/util/tempfiles"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
)

const (
	Version = "0.2.0"

	// DOLTGRES_DATA_DIR is an environment variable that defines the location of DoltgreSQL databases
	DOLTGRES_DATA_DIR = "DOLTGRES_DATA_DIR"
	// DOLTGRES_DATA_DIR_DEFAULT is the portion to append to the user's home directory if DOLTGRES_DATA_DIR has not been specified
	DOLTGRES_DATA_DIR_DEFAULT = "doltgres/databases"
	// DOLTGRES_DATA_DIR_CWD is an environment variable that causes DoltgreSQL to use the current directory for the
	// location of DoltgreSQL databases, rather than the DOLTGRES_DATA_DIR. This means that it has priority over
	// DOLTGRES_DATA_DIR.
	DOLTGRES_DATA_DIR_CWD = "DOLTGRES_DATA_DIR_CWD"
)

var sqlServerDocs = cli.CommandDocumentationContent{
	ShortDesc: "Start a PostgreSQL-compatible server.",
	LongDesc: "By default, starts a PostgreSQL-compatible server on the dolt database in the current directory. " +
		"Databases are named after the directories they appear in" +
		"Parameters can be specified using a yaml configuration file passed to the server via " +
		"{{.EmphasisLeft}}--config <file>{{.EmphasisRight}}, or by using the supported switches and flags to configure " +
		"the server directly on the command line. If {{.EmphasisLeft}}--config <file>{{.EmphasisRight}} is provided all" +
		" other command line arguments are ignored.\n\nThis is an example yaml configuration file showing all supported" +
		" items and their default values:\n\n" +
		indentLines(sqlserver.ServerConfigAsYAMLConfig(sqlserver.DefaultServerConfig()).String()) + "\n\n" + `
SUPPORTED CONFIG FILE FIELDS:

{{.EmphasisLeft}}data_dir{{.EmphasisRight}}: A directory where the server will load dolt databases to serve, and create new ones. Defaults to the current directory.

{{.EmphasisLeft}}cfg_dir{{.EmphasisRight}}: A directory where the server will load and store non-database configuration data, such as permission information. Defaults {{.EmphasisLeft}}$data_dir/.doltcfg{{.EmphasisRight}}.

{{.EmphasisLeft}}log_level{{.EmphasisRight}}: Level of logging provided. Options are: {{.EmphasisLeft}}trace{{.EmphasisRight}}, {{.EmphasisLeft}}debug{{.EmphasisRight}}, {{.EmphasisLeft}}info{{.EmphasisRight}}, {{.EmphasisLeft}}warning{{.EmphasisRight}}, {{.EmphasisLeft}}error{{.EmphasisRight}}, and {{.EmphasisLeft}}fatal{{.EmphasisRight}}.

{{.EmphasisLeft}}privilege_file{{.EmphasisRight}}: "Path to a file to load and store users and grants. Defaults to {{.EmphasisLeft}}$doltcfg-dir/privileges.db{{.EmphasisRight}}. Will be created as needed.

{{.EmphasisLeft}}branch_control_file{{.EmphasisRight}}: Path to a file to load and store branch control permissions. Defaults to {{.EmphasisLeft}}$doltcfg-dir/branch_control.db{{.EmphasisRight}}. Will be created as needed.

{{.EmphasisLeft}}max_logged_query_len{{.EmphasisRight}}: If greater than zero, truncates query strings in logging to the number of characters given.

{{.EmphasisLeft}}behavior.read_only{{.EmphasisRight}}: If true database modification is disabled. Defaults to false.

{{.EmphasisLeft}}behavior.autocommit{{.EmphasisRight}}: If true every statement is committed automatically. Defaults to true. @@autocommit can also be specified in each session.

{{.EmphasisLeft}}behavior.dolt_transaction_commit{{.EmphasisRight}}: If true all SQL transaction commits will automatically create a Dolt commit, with a generated commit message. This is useful when a system working with Dolt wants to create versioned data, but doesn't want to directly use Dolt features such as dolt_commit(). 

{{.EmphasisLeft}}user.name{{.EmphasisRight}}: The username that connections should use for authentication

{{.EmphasisLeft}}user.password{{.EmphasisRight}}: The password that connections should use for authentication.

{{.EmphasisLeft}}listener.host{{.EmphasisRight}}: The host address that the server will run on.  This may be {{.EmphasisLeft}}localhost{{.EmphasisRight}} or an IPv4 or IPv6 address

{{.EmphasisLeft}}listener.port{{.EmphasisRight}}: The port that the server should listen on

{{.EmphasisLeft}}listener.max_connections{{.EmphasisRight}}: The number of simultaneous connections that the server will accept

{{.EmphasisLeft}}listener.read_timeout_millis{{.EmphasisRight}}: The number of milliseconds that the server will wait for a read operation

{{.EmphasisLeft}}listener.write_timeout_millis{{.EmphasisRight}}: The number of milliseconds that the server will wait for a write operation

{{.EmphasisLeft}}remotesapi.port{{.EmphasisRight}}: A port to listen for remote API operations on. If set to a positive integer, this server will accept connections from clients to clone, pull, etc. databases being served.

{{.EmphasisLeft}}user_session_vars{{.EmphasisRight}}: A map of user name to a map of session variables to set on connection for each session.

{{.EmphasisLeft}}cluster{{.EmphasisRight}}: Settings related to running this server in a replicated cluster. For information on setting these values, see https://docs.dolthub.com/sql-reference/server/replication

If a config file is not provided many of these settings may be configured on the command line.`,
	Synopsis: []string{
		"--config {{.LessThan}}file{{.GreaterThan}}",
		"[-H {{.LessThan}}host{{.GreaterThan}}] [-P {{.LessThan}}port{{.GreaterThan}}] [-u {{.LessThan}}user{{.GreaterThan}}] [-p {{.LessThan}}password{{.GreaterThan}}] [-t {{.LessThan}}timeout{{.GreaterThan}}] [-l {{.LessThan}}loglevel{{.GreaterThan}}] [--data-dir {{.LessThan}}directory{{.GreaterThan}}] [-r]",
	},
}

func indentLines(s string) string {
	sb := strings.Builder{}
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		sb.WriteRune('\t')
		sb.WriteString(line)
		sb.WriteRune('\n')
	}
	return sb.String()
}

func init() {
	server.DefaultProtocolListenerFunc = NewListener
	sqlserver.ExternalDisableUsers = true
	dfunctions.VersionString = Version
}

// RunOnDisk starts the server based on the given args, while also using the local disk as the backing store.
// The returned WaitGroup may be used to wait for the server to close.
func RunOnDisk(ctx context.Context, args []string, dEnv *env.DoltEnv) (*svcs.Controller, error) {
	return runServer(ctx, args, dEnv)
}

// RunInMemory starts the server based on the given args, while also using RAM as the backing store.
// The returned WaitGroup may be used to wait for the server to close.
func RunInMemory(args []string) (*svcs.Controller, error) {
	ctx := context.Background()
	fs := filesys.EmptyInMemFS("")
	dEnv := env.Load(ctx, env.GetCurrentUserHomeDir, fs, doltdb.InMemDoltDB, Version)
	globalConfig, _ := dEnv.Config.GetConfig(env.GlobalConfig)
	if globalConfig.GetStringOrDefault(config.UserNameKey, "") == "" {
		globalConfig.SetStrings(map[string]string{
			config.UserNameKey:  "postgres",
			config.UserEmailKey: "postgres@somewhere.com",
		})
	}

	return runServer(ctx, args, dEnv)
}

// runServer starts the server based on the given args, using the provided file system as the backing store.
// The returned WaitGroup may be used to wait for the server to close.
func runServer(ctx context.Context, args []string, dEnv *env.DoltEnv) (*svcs.Controller, error) {
	sqlServerCmd := sqlserver.SqlServerCmd{}
	if serverArgs, err := sqlServerCmd.ArgParser().Parse(append([]string{"sql-server"}, args...)); err == nil {
		if _, ok := serverArgs.GetValue("port"); !ok {
			args = append(args, "--port=5432")
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

	defer tempfiles.MovableTempFileProvider.Clean()

	err := dsess.InitPersistedSystemVars(dEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to load persisted system variables: %w", err)
	}

	ap := sqlserver.SqlServerCmd{}.ArgParser()
	help, _ := cli.HelpAndUsagePrinters(cli.CommandDocsForCommandString("sql-server", sqlServerDocs, ap))

	serverConfig, err := sqlserver.ServerConfigFromArgs(ap, help, args, dEnv)
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

	// We need a username and password for many SQL commands, so set defaults if they don't exist
	dEnv.Config.SetFailsafes(map[string]string{
		config.UserNameKey:  "postgres",
		config.UserEmailKey: "postgres@somewhere.com",
	})

	// Automatically initialize a doltgres database if necessary
	if !dEnv.HasDoltDir() {
		// Need to make sure that there isn't a doltgres item in the path.
		if exists, isDirectory := dEnv.FS.Exists("doltgres"); !exists {
			err := dEnv.FS.MkDirs("doltgres")
			if err != nil {
				return nil, err
			}
			subdirectoryFS, err := dEnv.FS.WithWorkingDir("doltgres")
			if err != nil {
				return nil, err
			}

			// We'll use a temporary environment to instantiate the subdirectory
			tempDEnv := env.Load(ctx, env.GetCurrentUserHomeDir, subdirectoryFS, dEnv.UrlStr(), Version)
			res := commands.InitCmd{}.Exec(ctx, "init", []string{}, tempDEnv, configCliContext{tempDEnv})
			if res != 0 {
				return nil, fmt.Errorf("failed to initialize doltgres database")
			}
		} else if !isDirectory {
			workingDir, _ := dEnv.FS.Abs(".")
			// The else branch means that there's a Doltgres item, so we need to error if it's a file since we
			// enforce the creation of a Doltgres database/directory, which would create a name conflict with the file
			return nil, fmt.Errorf("Attempted to create the default `doltgres` database at `%s`, but a file with "+
				"the same name was found. Either remove the file, or change the environment variable `%s` so that it "+
				"points to a different directory.", workingDir, DOLTGRES_DATA_DIR)
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

	sqlserver.ConfigureServices(serverConfig, controller, Version, dEnv)
	go controller.Start(newCtx)
	return controller, controller.WaitForStart()
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
