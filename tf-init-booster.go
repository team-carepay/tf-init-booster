package main

import (
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"log"
	"os/user"
	"path/filepath"
	"strings"
)

func main() {
	if usr, err := user.Current(); err != nil {
		log.Fatal(err)
	} else if modules, err := ScanModules(); err != nil {
		log.Fatal(err)
	} else if err := CopyModules(modules, filepath.Join(usr.HomeDir, ".terraform.d/repositories"), GetAuth); err != nil {
		log.Fatal(err)
	} else if err := WriteModules(modules, ".terraform/modules/modules.json"); err != nil {
		log.Fatal(err)
	}
}

func GetAuth(host string) (transport.AuthMethod, error) {
	if privateKeyFile, err := ExpandFileName(ssh.DefaultSSHConfig.Get(host, "IdentityFile")); err != nil {
		return nil, err
	} else if publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKeyFile, ""); err != nil {
		return nil, err
	} else {
		return publicKeys, nil
	}
}

func ExpandFileName(filename string) (string, error) {
	if strings.HasPrefix(filename, "~/") {
		return HomeDirFileName(filename[2:])
	} else {
		return filename, nil
	}
}

func HomeDirFileName(filename string) (string, error) {
	if usr, err := user.Current(); err != nil {
		return "", err
	} else {
		return filepath.Join(usr.HomeDir, filename), nil
	}
}
