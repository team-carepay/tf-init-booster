package main

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type repoKey struct {
	host string
	path string
}

type Scanner struct {
	cacheDir     string
	repositories map[repoKey]*Repository
	checkouts    map[checkoutKey]*Checkout
	modules      map[string]*ModuleRef
}

type checkoutKey struct {
	host   string
	path   string
	branch string
}

func NewScanner(cacheDir string) *Scanner {
	scanner := Scanner{
		cacheDir: cacheDir,
	}
	return &scanner
}

func ScanModules() ([]*ModuleRef, error) {
	moduleRefRegex := regexp.MustCompile(`(?s)module\s*"([a-zA-Z0-9_\.-]+)"\s*{.*?source\s*=\s*"git@([a-zA-Z0-9_\.-]+):([a-zA-Z0-9_\/-]+)\.git(\/\/[a-zA-Z0-9_\/-]+)?(\?ref=[a-zA-Z0-9_\/\.-]+)?"`)
	var modules []*ModuleRef
	err := filepath.Walk(".",
		func(file string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && (info.Name() == ".terraform" || info.Name() == ".git") {
				return filepath.SkipDir
			}
			if !info.IsDir() && filepath.Ext(file) == ".tf" {
				if content, err := ioutil.ReadFile(file); err == nil {
					matches := moduleRefRegex.FindAllStringSubmatch(string(content), -1)
					for _, match := range matches {
						name, host, path, submodule, branch := match[1], match[2], match[3], match[4], match[5]
						checkoutDir := filepath.Join(".terraform/modules", name)
						modules = append(modules, NewModuleRef(name, host, path, submodule, branch, checkoutDir))
					}
				}
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return modules, nil
}

func CopyModules(modules []*ModuleRef, cacheDir string) error {
	repositories := make(map[repoKey]*Repository)
	checkouts := make(map[checkoutKey]*Checkout)
	for _, m := range modules {
		repokey := repoKey{host: m.Host, path: m.Path}
		repository, ok := repositories[repokey]
		if !ok {
			auth, err := GetAuth(m.Host)
			if err != nil {
				return err
			}
			repository = NewRepository(m.Host, m.Path, filepath.Join(cacheDir, m.Host, m.Path))
			if err := repository.Fetch(auth); err != nil {
				return err
			}
			repositories[repokey] = repository
		}
		checkoutkey := checkoutKey{host: m.Host, path: m.Path, branch: m.Branch}
		checkout, found := checkouts[checkoutkey]
		if !found {
			checkout = NewCheckout(repository, m.Branch, m.Dir)
			if err := checkout.Copy(); err != nil {
				return err
			}
			checkouts[checkoutkey] = checkout
		} else {
			if err := os.Symlink(filepath.Base(checkout.Dir), m.Dir); err != nil && err != os.ErrExist {
				return err
			}
		}
	}
	return nil
}

func GetAuth(host string) (*ssh.PublicKeys, error) {
	if privateKeyFile, err := ExpandFileName(ssh.DefaultSSHConfig.Get(host, "IdentityFile")); err != nil {
		return nil, err
	} else {
		return ssh.NewPublicKeysFromFile("git", privateKeyFile, "")
	}
}

func ExpandFileName(filename string) (string, error) {
	if strings.HasPrefix(filename, "~/") {
		return homeDirFileName(filename[2:])
	} else {
		return filename, nil
	}
}

func homeDirFileName(filename string) (string, error) {
	if usr, err := user.Current(); err != nil {
		return "", err
	} else {
		return filepath.Join(usr.HomeDir, filename), nil
	}
}
