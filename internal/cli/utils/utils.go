package utils

import (
	"fmt"
	"os"
	"preq/internal/cli/paramutils"
	"preq/internal/gitutils"
	"preq/internal/persistance"
	"preq/internal/pkg/client"
	"preq/internal/systemcodes"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func HasRepoFlagsSet(flags paramutils.FlagRepo) bool {
	repoFlag := flags.GetStringOrDefault("repository", "")
	providerFlag := flags.GetStringOrDefault("provider", "")
	return repoFlag != "" && providerFlag != ""
}

type ProcessPullRequestResponse struct {
	ID       string
	GlobalID string
	Status   string
	Error    error
}

func maxPRDescriptionLength(prs []*client.PullRequest, limit int) int {
	maxLen := 0
	for _, pr := range prs {
		l := len(pr.Source.Name) + len(pr.Destination.Name) + 4
		if l > maxLen {
			maxLen = l
		}
	}

	if limit > 0 && maxLen > limit {
		return limit
	}

	return maxLen
}

type (
	runCommandError   func(*cobra.Command, []string) error
	runCommandNoError func(*cobra.Command, []string)
)

func RunCommandWrapper(fn runCommandError) runCommandNoError {
	return func(cmd *cobra.Command, args []string) {
		err := fn(cmd, args)
		if err != nil {
			fmt.Println(err)

			switch err {
			// TODO: Add more specific exit codes
			default:
				os.Exit(systemcodes.ErrorCodeGeneric)
			}
		}
	}
}

func SafelyWriteVisitToState(
	flags *pflag.FlagSet,
	params *paramutils.RepositoryParams,
) {
	repoFlag, _ := flags.GetString("repository")
	providerFlag, _ := flags.GetString("provider")
	isExplicitRepo := repoFlag != "" && providerFlag != ""

	// Explicitly defiend repos should not be stored as the dir
	// cannot be infered from the working directory
	if isExplicitRepo {
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Error().Err(err).Msg("unabled to get the current working directory when writing to state")
		return
	}

	repoDir, err := gitutils.GetRepoRootDir(wd)
	if err != nil {
		return
	}

	err = persistance.GetDefault().AddVisited(
		params.Name,
		string(params.Provider),
		repoDir,
	)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
}
