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
	"os"
	"strconv"
	"strings"

	"github.com/dolthub/dolt/go/cmd/dolt/cli"
	"github.com/dolthub/dolt/go/cmd/dolt/commands"
	"github.com/dolthub/dolt/go/cmd/dolt/commands/sqlserver"
	eventsapi "github.com/dolthub/dolt/go/gen/proto/dolt/services/eventsapi/v1alpha1"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dfunctions"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/events"
	"github.com/dolthub/dolt/go/libraries/utils/config"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/dolthub/dolt/go/store/nbs"
	"github.com/dolthub/dolt/go/store/util/tempfiles"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/fatih/color"
)

const (
	Version = "0.1.0"
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
	events.Application = eventsapi.AppID_APP_DOLTGRES 
}

const chdirFlag = "--chdir"
const stdInFlag = "--stdin"
const stdOutFlag = "--stdout"
const stdErrFlag = "--stderr"
const stdOutAndErrFlag = "--out-and-err"

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
	dEnv := env.Load(ctx, env.GetCurrentUserHomeDir, fs, doltdb.LocalDirDoltDB, Version)
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

	if os.Getenv("DOLT_VERBOSE_ASSERT_TABLE_FILES_CLOSED") == "" {
		nbs.TableIndexGCFinalizerWithStackTrace = false
	}
	
	args, err := redirectStdio(args)
	if err != nil {
		return nil, err
	}
	
	if dEnv.CfgLoadErr != nil {
		return nil, fmt.Errorf("failed to load the global config: %w", dEnv.CfgLoadErr)
	}

	if dEnv.HasDoltDataDir() {
		return nil, fmt.Errorf("Cannot start a server within a directory containing a Dolt or Doltgres database." +
				"To use the current directory as a database, start the server from the parent directory.")
	}

	defer tempfiles.MovableTempFileProvider.Clean()

	err = dsess.InitPersistedSystemVars(dEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to load persisted system variables: %w", err)
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

	// We need a username and password for many SQL commands, so set defaults if they don't exist
	dEnv.Config.SetFailsafes(map[string]string{
		config.UserNameKey:  "postgres",
		config.UserEmailKey: "postgres@somewhere.com",
	})

	// Automatically initialize a doltgres database if necessary
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
			res := commands.InitCmd{}.Exec(ctx, "init", []string{}, tempDEnv, nil)
			if res != 0 {
				return nil, fmt.Errorf("failed to initialize doltgres database")
			}
		}
	}

	// If we got this far, emit a usage event in the background while we launch the server
	// Dolt is more permissive with events: it emits events even if the command fails in earliest phase.
	// We'll also emit a heartbeat event every 24 hours the server is running. All events will be tagged with the
	// doltgresql app id.
	go emitUsageEvent(dEnv)

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

func emitUsageEvent(dEnv *env.DoltEnv) {
	metricsDisabled := dEnv.Config.GetStringOrDefault(config.MetricsDisabled, "false")
	disabled, err := strconv.ParseBool(metricsDisabled)
	if err != nil || disabled {
		return
	}

	evt := events.NewEvent(sqlserver.SqlServerCmd{}.EventType())
	evtCollector := events.NewCollector()
	evtCollector.CloseEventAndAdd(evt)
	clientEvents := evtCollector.Close()
	
	emitter, err := commands.GRPCEmitterForConfig(dEnv)
	if err != nil {
		return
	}
	
	err = emitter.LogEvents(Version, clientEvents)
}