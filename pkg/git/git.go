// pkg/git/git.go
package git

import (
	"fmt"

	git "github.com/go-git/go-git/v5"
)

// GetStagedFiles returns a list of files staged in the current Git repository
func GetStagedFiles() ([]string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to access worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository status: %w", err)
	}

	var files []string
	for file := range status {
		files = append(files, file)
	}

	return files, nil
}
