package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"

	sshserver "github.com/gliderlabs/ssh"
	transportssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh"
)

var publicKeys *transportssh.PublicKeys

func TestAll(t *testing.T) {
	generateKeyPair()
	if modules, err := ScanModules(); err != nil {
		t.Error(err)
	} else {
		if len(modules) != 1 {
			t.Errorf("Expected size 1")
		}
		transportssh.DefaultSSHConfig = &mockSSHConfig{map[string]map[string]string{
			"bitbucket.org": {
				"Hostname": "localhost",
				"Port":     "2222",
			},
		}}
		log.Println("starting ssh server on port 2222...")
		server := &sshserver.Server{
			Addr: "localhost:2222",
			Handler: func(s sshserver.Session) {
				log.Printf("New session\n")
				cmd := exec.Command("git-upload-pack", ".")
				cmd.Stdin = s
				cmd.Stdout = s
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					log.Printf("command failed\n")
					t.Error(err)
				}
			},
		}
		defer server.Close()
		go func() { _ = server.ListenAndServe() }()

		if err := CopyModules(modules, "cache", testKeys); err != nil {
			log.Printf("copy failed\n")
			t.Error(err)
		}
		if err := WriteModules(modules, "modules.json"); err != nil {
			t.Error(err)
		}
		content, err := ioutil.ReadFile("modules.json")
		if err != nil {
			t.Error(err)
		}
		var m Modules
		if err := json.Unmarshal(content, &m); err != nil {
			t.Error(err)
		}
		if len(m.Modules) != 1 {
			t.Error("Expected one element")
		}
		m1 := m.Modules[0]
		if m1.Dir != ".terraform/modules/edge-router/edge-router" {
			t.Errorf("Wrong dir: %s", m1.Dir)
		}
		if m1.Source != "git@bitbucket.org:carepaydev/ssi-platform-modules.git//edge-router?ref=edge-router_1.0.5" {
			t.Errorf("Wrong source: %s", m1.Source)
		}
		if m1.Key != "edge-router" {
			t.Errorf("Wrong Key: %s", m1.Key)
		}
		// remove .terraform modules, run again
		os.RemoveAll(".terraform") // remove modules, forcing a rerun
		if err := CopyModules(modules, "cache", testKeys); err != nil {
			log.Printf("copy failed\n")
			t.Error(err)
		}
		if err := WriteModules(modules, "modules.json"); err != nil {
			t.Error(err)
		}
		// one more time with .terraform folder present (check idempotent)
		if err := CopyModules(modules, "cache", testKeys); err != nil {
			log.Printf("copy failed\n")
			t.Error(err)
		}
		if err := WriteModules(modules, "modules.json"); err != nil {
			t.Error(err)
		}
	}
}

func generateKeyPair() {
	if privateKey, err := rsa.GenerateKey(rand.Reader, 4096); err == nil {
		privDER := x509.MarshalPKCS1PrivateKey(privateKey)
		privBlock := pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privDER,
		}
		privatePEM := pem.EncodeToMemory(&privBlock)
		if publicKeys, err = transportssh.NewPublicKeys("git", privatePEM, ""); err == nil {
			publicKeys.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		}
	}
}

func testKeys(host string) (transport.AuthMethod, error) {
	return publicKeys, nil
}

type mockSSHConfig struct {
	Values map[string]map[string]string
}

func (c *mockSSHConfig) Get(alias, key string) string {
	a, ok := c.Values[alias]
	if !ok {
		return c.Values["*"][key]
	}

	return a[key]
}
