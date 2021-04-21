package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type Modules struct {
	Modules []Module `json:"Modules"`
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

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	//if _, err := os.Stat(".terraform/modules/modules.json"); err == nil {
	//	fmt.Printf("modules.json already exists, skipping")
	//	return
	//}
	gitCryptKey := os.Getenv("GIT_CRYPT_KEY")
	if err := os.MkdirAll(".terraform/modules", 0755); err != nil {
		log.Fatal(err)
	}
	r := regexp.MustCompile(`(?s)module\s*"([a-zA-Z0-9_\.-]+)"\s*{.*?source\s*=\s*"git@([a-zA-Z0-9_\.-]+):([a-zA-Z0-9_\/-]+)\.git(\/\/[a-zA-Z0-9_\/-]+)?(\?ref=[a-zA-Z0-9_\/\.-]+)?"`)
	repos := map[repoKey]string{}
	modules := map[string]Module{"": {Key: "", Source: "", Dir: "."}}
	branches := map[repoKey]map[string]string{}
	err = filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && (info.Name() == ".terraform" || info.Name() == ".git") {
				return filepath.SkipDir
			}
			if !info.IsDir() && filepath.Ext(path) == ".tf" {
				if content, err := ioutil.ReadFile(path); err == nil {
					matches := r.FindAllStringSubmatch(string(content), -1)
					for _, match := range matches {
						name, host, repopath, submodule, branch := match[1], match[2], match[3], match[4], match[5]
						var branchName string
						if branch == "" {
							branchName = "refs/heads/master" // TODO
						} else {
							branchName = "refs/tags/" + branch[5:]
						}
						key := repoKey{host: host, path: repopath}
						workdir := filepath.Join(usr.HomeDir, ".terraform.d/repositories", host, repopath)
						moduleDir := filepath.Join(".terraform/modules", name)
						_, ok := repos[key]
						if !ok {
							branches[key] = map[string]string{}
							auth, err := getAuth(host)
							if err != nil {
								return err
							}
							if _, err := os.Stat(workdir); os.IsNotExist(err) {
								if _, err = git.PlainClone(workdir, false, &git.CloneOptions{
									URL:  fmt.Sprintf("git@%s:%s.git", host, repopath),
									Auth: auth,
								}); err != nil {
									return err
								}
							} else if repo, err := git.PlainOpen(workdir); err != nil {
								return err
							} else if err := repo.Fetch(&git.FetchOptions{
								Auth: auth,
							}); err != nil && err != git.NoErrAlreadyUpToDate {
								return err
							}
							repos[key] = workdir
						}

						if branchDir, ok := branches[key][branch]; !ok {
							if err := copyDir(workdir, moduleDir); err != nil {
								return err
							} else if repo, err := git.PlainOpen(moduleDir); err != nil {
								return err
							} else if worktree, err := repo.Worktree(); err != nil {
								return err
							} else if err := worktree.Checkout(&git.CheckoutOptions{Branch: plumbing.ReferenceName(branchName)}); err != nil {
								return err
							}
							if gitCryptKey != "" {
								cmd := exec.Command("git-crypt", "unlock", gitCryptKey)
								cmd.Dir = moduleDir
								cmd.Stderr = os.Stderr
								if err := cmd.Run(); err != nil {
									return err
								}
							}
							branches[key][branch] = moduleDir
						} else {
							_ = os.RemoveAll(moduleDir)
							if err := os.Symlink(filepath.Base(branchDir), moduleDir); err != nil && !os.IsNotExist(err) {
								return err
							}
						}
						if strings.HasPrefix(submodule, "//") {
							moduleDir = filepath.Join(moduleDir, submodule[2:])
						}
						modules[name] = Module{
							Key:    name,
							Source: fmt.Sprintf("git@%s:%s.git%s%s", host, repopath, submodule, branch),
							Dir:    moduleDir,
						}
					}
				}
			}
			return nil
		})
	if err != nil {
		log.Fatal(err)
	}
	var moduleArray []Module
	for _, v := range modules {
		moduleArray = append(moduleArray, v)
	}
	if modulesJson, err := json.Marshal(Modules{Modules: moduleArray}); err != nil {
		log.Fatal(err)
	} else {
		if err := ioutil.WriteFile(".terraform/modules/modules.json", modulesJson, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func copyDir(scrDir, dest string) error {
	if entries, err := ioutil.ReadDir(scrDir); err != nil {
		return err
	} else {
		for _, entry := range entries {
			sourcePath := filepath.Join(scrDir, entry.Name())
			destPath := filepath.Join(dest, entry.Name())

			switch entry.Mode() & os.ModeType {
			case os.ModeDir:
				if err := createIfNotExists(destPath, entry.Mode()); err != nil {
					return err
				}
				if err := copyDir(sourcePath, destPath); err != nil {
					return err
				}
			case os.ModeSymlink:
				if err := copySymLink(sourcePath, destPath); err != nil {
					return err
				}
			default:
				if err := copyFile(sourcePath, destPath, entry.Mode()); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func copyFile(srcFile, dstFile string, mode os.FileMode) error {
	if out, err := os.OpenFile(dstFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode); err != nil {
		return err
	} else {
		defer out.Close()
		if in, err := os.Open(srcFile); err != nil {
			return err
		} else {
			defer in.Close()
			if _, err = io.Copy(out, in); err != nil {
				return err
			}
			return nil
		}
	}
}

func exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func createIfNotExists(dir string, perm os.FileMode) error {
	if exists(dir) {
		return nil
	}
	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}
	return nil
}

func copySymLink(source, dest string) error {
	if link, err := os.Readlink(source); err != nil {
		return err
	} else {
		if _, err := os.Lstat(dest); err == nil {
			_ = os.Remove(dest)
		}
		if err := os.Symlink(link, dest); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
}

func getAuth(host string) (*ssh.PublicKeys, error) {
	if privateKeyFile, err := expandFileName(ssh.DefaultSSHConfig.Get(host, "IdentityFile")); err != nil {
		return nil, err
	} else {
		return ssh.NewPublicKeysFromFile("git", privateKeyFile, "")
	}
}

func expandFileName(filename string) (string, error) {
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
