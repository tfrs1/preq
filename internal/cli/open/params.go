package open

import (
	"preq/internal/cli/paramutils"
)

type openCmdParams struct {
	Repository paramutils.RepositoryParams
	PrintOnly  bool
}

type cmdArgs struct {
	ID string
}

func parseArgs(args []string) *cmdArgs {
	return &cmdArgs{ID: paramutils.ParseIDArg(args)}
}

func fillDefaultOpenCmdParams(params *openCmdParams) {
	params.PrintOnly = false
}

func fillFlagOpenCmdParams(
	flags *paramutils.PFlagSetWrapper,
	params *openCmdParams,
) {
	params.PrintOnly = flags.GetBoolOrDefault("print", false)
}
