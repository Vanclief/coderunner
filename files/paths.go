package files

import (
	"fmt"

	"github.com/vanclief/coderunner/git"
	"github.com/vanclief/ez"
)

func GetScopeFilePath(scope string) (string, error) {
	const op = "files.GetScopeFilePath"

	return GetCommitFilePath(scope, "json")
}

func GetCommitFilePath(fileName, fileExtension string) (string, error) {
	const op = "files.GetFilePath"

	gitInfo, err := git.GetInfo()
	if err != nil {
		return "", ez.Wrap(op, err)
	}

	return fmt.Sprintf("%s/%s.%s.%s", CODERUNNER_DIR, gitInfo.CurrentCommit, fileName, fileExtension), nil
}
