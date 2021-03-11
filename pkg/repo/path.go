package repo

import "github.com/mitchellh/go-homedir"

const defaultRepoDir = "~/.venus-messager"

func GetRepoPath(override string) (string, error) {
	// override is first precedence
	if override != "" {
		return homedir.Expand(override)
	}
	// Default is third precedence
	return homedir.Expand(defaultRepoDir)
}
