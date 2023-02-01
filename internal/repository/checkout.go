package repository

import (
	"fmt"
	"os"
	"os/exec"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	dirutil "github.com/team-carepay/tf-init-booster/internal/dirutil"
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
	needsUnlock := false
	needsLock := false
	fileInfo, err := os.Lstat(checkout.Dir)
	if err != nil || fileInfo.Mode()&os.ModeSymlink != 0 {
		_ = os.RemoveAll(checkout.Dir)
		if err := dirutil.CopyDir(checkout.Repository.Dir, checkout.Dir); err != nil {
			return err
		}
		needsUnlock = true
	}

	repo, err := git.PlainOpen(checkout.Dir)
	if err != nil {
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	head, err := repo.Head()
	if err != nil {
		return err
	}
	target, err := repo.ResolveRevision(plumbing.Revision(checkout.Ref))
	if err != nil {
		return fmt.Errorf("unable to find branch of tag %s", checkout.Ref)
	}
	if head.Hash() != *target {
		if err := worktree.Reset(&git.ResetOptions{
			Commit: *target,
			Mode:   git.HardReset,
		}); err != nil {
			return err
		}
		needsLock = true
		needsUnlock = true
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
