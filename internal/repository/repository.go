package repository

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	dirutil "github.com/team-carepay/tf-init-booster/internal/dirutil"

	git "github.com/go-git/go-git/v5"
)

type Auth interface {
	AuthToken()
	AuthSSH(host string)
	Get() (transport.AuthMethod, error)
}

type GetAuth struct {
	Auth transport.AuthMethod
	Err  error
	Repository
}

func (g *GetAuth) Get() (transport.AuthMethod, error) {
	var (
		accessType string = os.Getenv("REPO_ACCESS")
	)

	if accessType == "token" {
		g.AuthToken()
		if g.Err != nil {
			return nil, fmt.Errorf("unable to get auth token, please check if the token exist")
		}
		return g.Auth, nil
	}

	if accessType == "ssh" || accessType == "" {
		g.AuthSSH(g.Host)

		if g.Err != nil {
			return nil, fmt.Errorf("unable to get auth ssh, please check if the ssh key exist")
		}
		return g.Auth, nil
	}

	return nil, fmt.Errorf("unable to get auth, please check if your config is correct")
}

func (g *GetAuth) AuthToken() {
	tokenProvider := os.Getenv("TF_TOKEN_PROVIDER")
	accessUser := os.Getenv("TF_ACCESS_USER")
	accessToken := os.Getenv(tokenProvider)

	if accessToken == "" || tokenProvider == "" {
		g.Auth = nil
		g.Err = fmt.Errorf("unable to get auth token please check if the variable exist")
		g.Url = fmt.Sprintf("https://%s/%s.git", g.Host, g.Path)
	}
	access := &http.BasicAuth{
		Username: accessUser,
		Password: accessToken,
	}
	g.Auth = access
}

func (g *GetAuth) AuthSSH(host string) {
	privateKeyFile, err := dirutil.ExpandFileName(ssh.DefaultSSHConfig.Get(host, "IdentityFile"))
	if err != nil {
		g.Auth = nil
		g.Err = err
		g.Url = fmt.Sprintf("git@%s:%s.git", g.Host, g.Path)
	}
	publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKeyFile, "")
	if err != nil {
		g.Auth = nil
		g.Err = err
	}
	g.Auth = publicKeys
}

type Repository struct {
	Host string
	Path string
	Dir  string
	Auth transport.AuthMethod
	Url  string
}

func NewRepository(host, path, dir string) *Repository {
	return &Repository{
		Host: host,
		Path: path,
		Dir:  dir,
	}
}

func (r *Repository) Fetch(auth transport.AuthMethod) error {
	if _, err := os.Stat(r.Dir); os.IsNotExist(err) {
		if _, err = git.PlainClone(r.Dir, false, &git.CloneOptions{
			URL:  r.Url,
			Auth: auth,
		}); err != nil {
			return err
		}
	} else if repo, err := git.PlainOpen(r.Dir); err != nil {
		return err
	} else if err := repo.Fetch(&git.FetchOptions{
		Auth: auth,
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	return nil
}
