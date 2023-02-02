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
	AuthToken() (transport.AuthMethod, error)
	AuthSSH(host string) (transport.AuthMethod, error)
	Get() (transport.AuthMethod, string, error)
}

type GetAuth struct {
	Auth transport.AuthMethod
	Err  error
	Repository
}

func (g *GetAuth) Get() (transport.AuthMethod, string, error) {
	var (
		accessType string = os.Getenv("REPO_ACCESS")
	)

	if accessType == "token" {
		auth, err := g.AuthToken()
		if err != nil {
			return nil, "", err
		}
		url := fmt.Sprintf("https://%s/%s", g.Host, g.Path)
		return auth, url, nil
	}
	url := fmt.Sprintf("git@%s:%s.git", g.Host, g.Path)
	auth, err := g.AuthSSH(g.Host)
	if err != nil {
		return nil, "", err
	}
	return auth, url, nil
}

func (g *GetAuth) AuthToken() (transport.AuthMethod, error) {
	tokenProvider := os.Getenv("TF_TOKEN_PROVIDER")
	accessUser := os.Getenv("TF_ACCESS_USER")
	accessToken := os.Getenv(tokenProvider)

	if accessToken == "" || tokenProvider == "" {
		return nil, fmt.Errorf("unable to get auth token please check if the variable exist")
	}
	access := &http.BasicAuth{
		Username: accessUser,
		Password: accessToken,
	}
	return access, nil
}

func (g *GetAuth) AuthSSH(host string) (transport.AuthMethod, error) {
	privateKeyFile, err := dirutil.ExpandFileName(ssh.DefaultSSHConfig.Get(host, "IdentityFile"))
	if err != nil {
		return nil, err
	}
	publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKeyFile, "")
	if err != nil {
		return nil, err
	}
	return publicKeys, nil
}

type Repository struct {
	Host string
	Path string
	Dir  string
	Auth transport.AuthMethod
}

func NewRepository(host, path, dir string) *Repository {
	return &Repository{
		Host: host,
		Path: path,
		Dir:  dir,
	}
}

func (r *Repository) Fetch(auth transport.AuthMethod, url string) error {
	if _, err := os.Stat(r.Dir); os.IsNotExist(err) {
		if _, err = git.PlainClone(r.Dir, false, &git.CloneOptions{
			URL:  url,
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
