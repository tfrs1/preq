package decline

import (
	"preq/internal/cli/paramutils"
	"preq/internal/gitutils"
)

var getWorkingDirectoryRepo = gitutils.GetWorkingDirectoryRepo

type cmdParams struct {
	Repository paramutils.RepositoryParams
}

type cmdArgs struct {
	ID string
}

func parseArgs(args []string) *cmdArgs {
	id := ""
	if len(args) > 0 {
		id = args[0]
	}

	return &cmdArgs{ID: id}
}
