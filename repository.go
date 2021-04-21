package main

import (
	"fmt"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type Repository struct {
	Host string
	Path string
	Dir  string
}

func NewRepository(host, path, dir string) *Repository {
	return &Repository{
		Host: host,
		Path: path,
		Dir:  dir,
	}
}

func (repository *Repository) Fetch(auth *ssh.PublicKeys) error {
	if _, err := os.Stat(repository.Dir); os.IsNotExist(err) {
		if _, err = git.PlainClone(repository.Dir, false, &git.CloneOptions{
			URL:  fmt.Sprintf("git@%s:%s.git", repository.Host, repository.Path),
			Auth: auth,
		}); err != nil {
			return err
		}
	} else if repo, err := git.PlainOpen(repository.Dir); err != nil {
		return err
	} else if err := repo.Fetch(&git.FetchOptions{
		Auth: auth,
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	return nil
}
