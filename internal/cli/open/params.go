package open

import (
	"preq/internal/cli/paramutils"

	"github.com/spf13/cobra"
)

type openCmdParams struct {
	ID        string
	PrintOnly bool
}

func newOpenCmdParams() *openCmdParams {
	return &openCmdParams{ID: "", PrintOnly: false}
}

func fillArgParams(params *openCmdParams, args []string) {
	params.ID = paramutils.ParseIDArg(args)
}

func fillFlagOpenCmdParams(params *openCmdParams, cmd *cobra.Command) {
	var (
		flags     = &paramutils.PFlagSetWrapper{Flags: cmd.Flags()}
		printOnly = flags.GetBoolOrDefault("print", false)
	)

	params.PrintOnly = printOnly
}
