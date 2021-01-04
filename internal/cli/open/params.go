package open

import (
	"preq/internal/cli/paramutils"
	"preq/internal/config"

	"github.com/spf13/cobra"
)

type openCmdParams struct {
	Repository  config.RepositoryParams
	PrintOnly   bool
	Interactive bool
}

type cmdArgs struct {
	ID string
}

func parseArgs(args []string) *cmdArgs {
	return &cmdArgs{ID: paramutils.ParseIDArg(args)}
}

func fillDefaultOpenCmdParams(params *openCmdParams) {
	params.PrintOnly = false
	params.Interactive = false
	config.FillDefaultRepositoryParams(&params.Repository)
}

func fillFlagOpenCmdParams(cmd *cobra.Command, params *openCmdParams) {
	var (
		flags       = &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}
		printOnly   = flags.GetBoolOrDefault("print", false)
		interactive = flags.GetBoolOrDefault("interactive", false)
	)

	config.FillFlagRepositoryParams(flags, &params.Repository)
	params.PrintOnly = printOnly
	params.Interactive = interactive
}

func validateFlagOpenCmdParams(params *openCmdParams) error {
	return paramutils.ValidateFlagRepositoryParams(&params.Repository)
}
