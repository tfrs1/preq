package open

import (
	"preq/internal/cli/paramutils"

	"github.com/spf13/cobra"
)

type openCmdParams struct {
	Repository  paramutils.RepositoryParams
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
	paramutils.FillDefaultRepositoryParams(&params.Repository)
}

func fillFlagOpenCmdParams(cmd *cobra.Command, params *openCmdParams) {
	var (
		flags       = &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}
		printOnly   = flags.GetBoolOrDefault("print", false)
		interactive = flags.GetBoolOrDefault("interactive", false)
	)

	paramutils.FillFlagRepositoryParams(flags, &params.Repository)
	params.PrintOnly = printOnly
	params.Interactive = interactive
}

func validateFlagOpenCmdParams(params *openCmdParams) error {
	return paramutils.ValidateFlagRepositoryParams(&params.Repository)
}
