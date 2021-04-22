package main

import (
	"fmt"
	"os"
	"os/exec"

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

func (checkout *Checkout) Copy() error {
	gitCryptKey := os.Getenv("GIT_CRYPT_KEY")
	needsLock := false
	needsUnlock := false
	if fileInfo, err := os.Lstat(checkout.Dir); err != nil || fileInfo.Mode()&os.ModeSymlink != 0 {
		_ = os.RemoveAll(checkout.Dir)
		if err := CopyDir(checkout.Repository.Dir, checkout.Dir); err != nil {
			return err
		}
		needsUnlock = true
	} else {
		needsLock = true
	}

	if repo, err := git.PlainOpen(checkout.Dir); err != nil {
		return err
	} else if worktree, err := repo.Worktree(); err != nil {
		return err
	} else {
		head, err := repo.Head()
		if err != nil {
			return err
		}
		var target *plumbing.Reference
		tagRefs, err := repo.Tags()
		if err != nil {
			return err
		}
		err = tagRefs.ForEach(func(t *plumbing.Reference) error {
			if t.Name().String() == checkout.Ref {
				target = t
			}
			return nil
		})
		if target == nil {
			branchRefs, err := repo.Branches()
			if err != nil {
				return err
			}
			err = branchRefs.ForEach(func(t *plumbing.Reference) error {
				if t.Name().String() == checkout.Ref {
					target = t
				}
				return nil
			})
		}
		if target == nil {
			fmt.Errorf("Unable to find branch of tag %s", checkout.Ref)
		}
		if head.Hash() != target.Hash() {
			if err := worktree.Reset(&git.ResetOptions{
				Commit: target.Hash(),
				Mode: git.HardReset,
			}); err != nil {
				return err
			} else {
				needsLock = true
				needsUnlock = true
			}
		}
		if needsUnlock && gitCryptKey != "" {
			if needsLock {
				cmd := exec.Command("git-crypt", "lock", "--force")
				cmd.Dir = checkout.Dir
				_ = cmd.Run()
			}
			cmd := exec.Command("git-crypt", "unlock", gitCryptKey)
			cmd.Dir = checkout.Dir
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
			return nil
		}
		return nil
	}
}
