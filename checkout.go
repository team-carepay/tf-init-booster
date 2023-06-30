package main

import (
	"fmt"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Checkout struct {
	Repository *Repository
	Ref        string
	Dir        string
}

func NewCheckout(repository *Repository, ref, dir string) *Checkout {
	return &Checkout{
		Repository: repository,
		Ref:        ref,
		Dir:        dir,
	}
}

func (checkout *Checkout) Checkout() error {
	repo, err := git.PlainOpen(checkout.Repository.Dir)
	if err != nil {
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	target, err := repo.ResolveRevision(plumbing.Revision(checkout.Ref))
	if err != nil {
		return fmt.Errorf("unable to find branch of tag %s", checkout.Ref)
	}
	var st git.Status
	for i := 0; i < 10; i++ {
		if err := worktree.Checkout(&git.CheckoutOptions{
			Hash:  *target,
			Force: true,
		}); err != nil {
			return err
		}
		st, err = worktree.Status()
		if err != nil {
			return err
		}
		if st.IsClean() {
			return nil
		}
	}
	return fmt.Errorf("unable to checkout %s, working dir not clean: %+v", checkout.Ref, st)
}
