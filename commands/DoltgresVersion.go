package commands

import (
	"context"

	"github.com/dolthub/dolt/go/cmd/dolt/cli"
	"github.com/dolthub/dolt/go/cmd/dolt/commands"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/utils/argparser"
)

const doltVersionFlag = "dolt"

var versionDocs = cli.CommandDocumentationContent{
	ShortDesc: "Displays the version for the Doltgres binary.",
	LongDesc: `Displays the version for the Doltgres binary.

The out-of-date check can be disabled by running {{.EmphasisLeft}}doltgres config --global --add versioncheck.disabled true{{.EmphasisRight}}.`,
	Synopsis: []string{
		`[--verbose] [--feature] [--dolt]`,
	},
}

// DoltgresVersionCmd displays the version of the doltgres
type DoltgresVersionCmd struct {
	*commands.VersionCmd
	DoltVersionStr string
}

// Name is returns the name of the Dolt cli command. This is what is used on the command line to invoke the command
func (cmd DoltgresVersionCmd) Name() string {
	return "version"
}

// Description returns a description of the command
func (cmd DoltgresVersionCmd) Description() string {
	return versionDocs.ShortDesc
}

// RequiresRepo should return false if this interface is implemented, and the command does not have the requirement
// that it be run from within a data repository directory
func (cmd DoltgresVersionCmd) RequiresRepo() bool {
	return false
}

func (cmd DoltgresVersionCmd) Docs() *cli.CommandDocumentation {
	ap := cmd.ArgParser()
	return cli.NewCommandDocumentation(versionDocs, ap)
}

func (cmd DoltgresVersionCmd) ArgParser() *argparser.ArgParser {
	ap := cmd.VersionCmd.ArgParser()
	ap.SupportsFlag(doltVersionFlag, "d", "display the version of Dolt that this Doltgres binary is dependent on.")
	return ap
}

// Exec executes the command
func (cmd DoltgresVersionCmd) Exec(ctx context.Context, commandStr string, args []string, dEnv *env.DoltEnv, cliCtx cli.CliContext) int {
	ap := cmd.ArgParser()
	help, usage := cli.HelpAndUsagePrinters(cli.CommandDocsForCommandString(commandStr, versionDocs, ap))
	apr := cli.ParseArgsOrDie(ap, args, help)

	if apr.Contains(doltVersionFlag) {
		cli.Println("dolt version", cmd.DoltVersionStr)
		return 0
	} else {
		return cmd.VersionCmd.ExecWithArgParser(ctx, apr, usage, dEnv)
	}
}
