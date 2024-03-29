package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

type ModuleRef struct {
	Name      string
	Host      string
	Path      string
	Branch    string
	SubModule string
	Dir       string
	Ref       string
}

func NewModuleRef(name, host, path, submodule, branch, dir string) *ModuleRef {
	var ref string
	if branch == "" {
		ref = "master"
	} else {
		ref = branch[5:] // strip the prefix ?ref=
	}
	return &ModuleRef{
		Name:      name,
		Host:      host,
		Path:      path,
		Branch:    branch,
		SubModule: submodule,
		Dir:       dir,
		Ref:       ref,
	}
}

func (m *ModuleRef) ToModule() *Module {
	moduleDir := m.Dir
	if strings.HasPrefix(m.SubModule, "//") {
		moduleDir = filepath.Join(m.Dir, m.SubModule[2:])
	}

	return &Module{
		Key:    m.Name,
		Source: fmt.Sprintf("git@%s:%s.git%s%s", m.Host, m.Path, m.SubModule, m.Branch),
		Dir:    moduleDir,
	}
}

func WriteModules(modules map[string]*ModuleRef, file string) error {
	var moduleArray []*Module
	for _, v := range modules {
		moduleArray = append(moduleArray, v.ToModule())
	}
	if modulesJson, err := json.Marshal(Modules{Modules: moduleArray}); err != nil {
		return err
	} else {
		if err := ioutil.WriteFile(file, modulesJson, os.ModePerm); err != nil {
			return err
		} else {
			return nil
		}
	}
}

type Modules struct {
	Modules []*Module `json:"Modules"`
}

type Module struct {
	Key    string `json:"Key"`
	Source string `json:"Source"`
	Dir    string `json:"Dir"`
}

type repoKey struct {
	host string
	path string
}

type checkoutKey struct {
	host string
	path string
	ref  string
}

func ScanModules() (map[string]*ModuleRef, error) {
	moduleRefRegex := regexp.MustCompile(`(?s)module\s*"([a-zA-Z0-9_\.-]+)"\s*{.*?source\s*=\s*"git@([a-zA-Z0-9_\.-]+):([a-zA-Z0-9_\/-]+)\.git(\/\/[a-zA-Z0-9_\/-]+)?(\?ref=[a-zA-Z0-9_\/\.-]+)?"`)
	var modules = make(map[string]*ModuleRef)
	err := filepath.Walk(".",
		func(file string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && (info.Name() == ".terraform" || info.Name() == ".git") {
				return filepath.SkipDir
			}
			if !info.IsDir() && filepath.Ext(file) == ".tf" && !strings.HasSuffix(file, "override.tf") {
				if content, err := ioutil.ReadFile(file); err == nil {
					matches := moduleRefRegex.FindAllStringSubmatch(string(content), -1)
					for _, match := range matches {
						name, host, path, submodule, branch := match[1], match[2], match[3], match[4], match[5]
						checkoutDir := filepath.Join(".terraform/modules", name)
						modules[name] = NewModuleRef(name, host, path, submodule, branch, checkoutDir)
					}
				}
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	err = filepath.Walk(".",
		func(file string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && (info.Name() == ".terraform" || info.Name() == ".git") {
				return filepath.SkipDir
			}
			if !info.IsDir() && filepath.Ext(file) == ".tf" && strings.HasSuffix(file, "override.tf") {
				if content, err := ioutil.ReadFile(file); err == nil {
					matches := moduleRefRegex.FindAllStringSubmatch(string(content), -1)
					for _, match := range matches {
						name, host, path, submodule, branch := match[1], match[2], match[3], match[4], match[5]
						checkoutDir := filepath.Join(".terraform/modules", name)
						modules[name] = NewModuleRef(name, host, path, submodule, branch, checkoutDir)
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

func CopyModules(modules map[string]*ModuleRef, cacheDir string, authFunc func(string) (transport.AuthMethod, error)) error {
	repositories := make(map[repoKey]*Repository)
	checkouts := make(map[checkoutKey]*Checkout)
	copier, err := NewCopier()
	if err != nil {
		return err
	}
	for _, m := range modules {
		repokey := repoKey{host: m.Host, path: m.Path}
		checkoutkey := checkoutKey{host: m.Host, path: m.Path, ref: m.Ref}
		repository, ok := repositories[repokey]
		if !ok {
			repository = NewRepository(m.Host, m.Path, filepath.Join(cacheDir, m.Host, m.Path))
			if auth, err := authFunc(m.Host); err != nil {
				return err
			} else if err := repository.Fetch(auth); err != nil {
				return err
			}
			repositories[repokey] = repository
		}
		if checkout, found := checkouts[checkoutkey]; !found {
			checkout = NewCheckout(repository, m.Ref, m.Dir)
			if err := checkout.Checkout(); err != nil {
				return err
			} else {
				fileInfo, err := os.Lstat(checkout.Dir)
				if err != nil || fileInfo.Mode()&os.ModeSymlink != 0 {
					_ = os.RemoveAll(checkout.Dir) // make sure destination does not exist
				}

				if err := copier.CopyDir(repository.Dir, checkout.Dir); err != nil {
					return err
				}
			}
			checkouts[checkoutkey] = checkout
		} else {
			if _, err := os.Lstat(m.Dir); err == nil {
				_ = os.RemoveAll(m.Dir) // remove old symlink or dir if exists
			}
			if err := os.Symlink(filepath.Base(checkout.Dir), m.Dir); err != nil && !os.IsExist(err) {
				return err
			}
		}
	}
	return nil
}
