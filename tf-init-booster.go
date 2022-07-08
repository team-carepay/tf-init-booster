package main

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	modules, err := ScanModules()
	if err != nil {
		log.Fatal(err)
	}
	if len(modules) > 0 {
		if len(os.Args) == 2 {
			for _, m := range modules {
				m.Dir = os.Args[1]
			}
		} else {
			if err := CopyModules(modules, filepath.Join(usr.HomeDir, ".terraform.d/repositories"), GetAuth); err != nil {
				log.Fatal(err)
			}
		}
		if err := WriteModules(modules, ".terraform/modules/modules.json"); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("No modules found, skipping booster")
	}
}

func GetAuth(host string) (transport.AuthMethod, error) {
	privateKeyFile, err := ExpandFileName(ssh.DefaultSSHConfig.Get(host, "IdentityFile"))
	if err != nil {
		return nil, err
	}
	publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKeyFile, "")
	if err != nil {
		return nil, err
	}
	return publicKeys, nil
}

func ExpandFileName(filename string) (string, error) {
	if strings.HasPrefix(filename, "~/") {
		return HomeDirFileName(filename[2:])
	}
	return filename, nil
}

func HomeDirFileName(filename string) (string, error) {
	if usr, err := user.Current(); err != nil {
		return "", err
	} else {
		return filepath.Join(usr.HomeDir, filename), nil
	}
}
