package main

import (
	"os"
	"os/exec"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Checkout struct {
	Repository *Repository
	Branch     string
	Dir        string
}

func NewCheckout(repository *Repository, branch, dir string) *Checkout {
	return &Checkout{
		Repository: repository,
		Branch:     branch,
		Dir:        dir,
	}
}

func (checkout *Checkout) Copy() error {
	if err := CopyDir(checkout.Repository.Dir, checkout.Dir); err != nil {
		return err
	} else if repo, err := git.PlainOpen(checkout.Dir); err != nil {
		return err
	} else if worktree, err := repo.Worktree(); err != nil {
		return err
	} else if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(checkout.Branch),
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func (checkout *Checkout) Unlock(keyfile string) error {
	cmd := exec.Command("git-crypt", "unlock", keyfile)
	cmd.Dir = checkout.Dir
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
