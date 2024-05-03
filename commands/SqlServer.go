package commands

import (
	"context"
	"strconv"

	"github.com/dolthub/dolt/go/cmd/dolt/cli"
	"github.com/dolthub/dolt/go/cmd/dolt/commands"
	"github.com/dolthub/dolt/go/cmd/dolt/commands/sqlserver"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/events"
	"github.com/dolthub/dolt/go/libraries/utils/argparser"
	"github.com/dolthub/dolt/go/libraries/utils/config"
	"github.com/fatih/color"

	"github.com/dolthub/doltgresql/server"
)

var sqlServerDocs = cli.CommandDocumentationContent{
	ShortDesc: "",
	LongDesc:  "",
}

// SqlServerCmd displays the version of the doltgres
type SqlServerCmd struct {
}

// Name is returns the name of the Dolt cli command. This is what is used on the command line to invoke the command
func (cmd SqlServerCmd) Name() string {
	return "version"
}

// Description returns a description of the command
func (cmd SqlServerCmd) Description() string {
	return versionDocs.ShortDesc
}

// RequiresRepo should return false if this interface is implemented, and the command does not have the requirement
// that it be run from within a data repository directory
func (cmd SqlServerCmd) RequiresRepo() bool {
	return false
}

func (cmd SqlServerCmd) Docs() *cli.CommandDocumentation {
	ap := cmd.ArgParser()
	return cli.NewCommandDocumentation(versionDocs, ap)
}

func (cmd SqlServerCmd) ArgParser() *argparser.ArgParser {
	ap := argparser.NewArgParserWithMaxArgs(cmd.Name(), 0)
	ap.SupportsFlag(doltVersionFlag, "d", "display the version of Dolt that this Doltgres binary is dependent on.")
	return ap
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

// Exec executes the command
func (cmd SqlServerCmd) Exec(ctx context.Context, commandStr string, args []string, dEnv *env.DoltEnv, cliCtx cli.CliContext) int {
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
		cli.Println(color.RedString(err.Error()))
		return 1
	}

	err = controller.WaitForStop()
	if err != nil {
		cli.Println(color.RedString(err.Error()))
		return 1
	}

	return 0
}
