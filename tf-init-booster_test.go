package main_test

import (
	"testing"
)

func TestInit(t *testing.T) {
	// 	os.Chdir("test-data")
	// 	hostPrivateKey, err := ioutil.ReadFile("ssh/id_rsa")
	// 	if err != nil {
	// 		log.Fatal("failed to read private key")
	// 	}
	// 	hostPrivateKeySigner, err := ssh.ParsePrivateKey(hostPrivateKey)
	// 	if err != nil {
	// 		log.Fatal("failed to parse private key")
	// 	}

	// 	authorizedKeysBytes, err := ioutil.ReadFile("ssh/authorized-keys")
	// 	if err != nil {
	// 		log.Fatal("failed to load authorized_keys")
	// 	}
	// 	authorizedKeys := make(map[string]struct{})
	// 	for len(authorizedKeysBytes) > 0 {
	// 		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
	// 		if err != nil {
	// 			log.Fatal("failed to parse authorized_keys")
	// 		}
	// 		authorizedKeys[string(pubKey.Marshal())] = struct{}{}
	// 		authorizedKeysBytes = rest
	// 	}

	// 	sshLis, err := net.Listen("tcp", ":2222")
	// 	if err != nil {
	// 		log.Fatal("failed to listen ssh")
	// 	}

	// 	s := gitssh.Server{
	// 		RepoDir:            "test-repo",
	// 		Signer:             hostPrivateKeySigner,
	// 		PublicKeyCallback:  loggingPublicKeyCallback(authorizedKeys),
	// 		GitRequestTransfer: loggingGitRequestTransfer("/bin/git-shell"),
	// 	}
	// 	go func() {
	// 		// ready <- struct{}{}
	// 		if err := s.Serve(sshLis); err != gitssh.ErrServerClosed && err != nil {
	// 			fmt.Errorf("failed to serve: %v", err)
	// 		}
	// 	}()

	// 	tfInitBooster.main()

	// 	ctx, cancel := context.WithTimeout(context.Background(), 30)
	// 	defer cancel()

	// 	if err := s.Shutdown(ctx); err != nil {
	// 		fmt.Printf("failed to shutdown: %v\n", err)
	// 	}
	// }

	// func loggingPublicKeyCallback(authorizedKeys map[string]struct{}) gitssh.PublicKeyCallback {
	// 	return func(metadata ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	// 		session := hex.EncodeToString(metadata.SessionID())
	// 		var err error
	// 		if _, ok := authorizedKeys[string(key.Marshal())]; !ok {
	// 			err = errors.New("failed to authorize")
	// 			return nil, err
	// 		}
	// 		return &ssh.Permissions{
	// 			CriticalOptions: map[string]string{},
	// 			Extensions: map[string]string{
	// 				"SESSION": session,
	// 			},
	// 		}, nil
	// 	}
	// }
	// func loggingGitRequestTransfer(shellPath string) gitssh.GitRequestTransfer {
	// 	t := gitssh.LocalGitRequestTransfer(shellPath)
	// 	return func(ctx context.Context, ch ssh.Channel, req *ssh.Request, perms *ssh.Permissions, packCmd, repoPath string) error {
	// 		startTime := time.Now()
	// 		chs := &ChannelWithSize{
	// 			Channel: ch,
	// 		}
	// 		var err error
	// 		defer func() {
	// 			finishTime := time.Now()

	// 			payload := string(req.Payload)
	// 			i := strings.Index(payload, "git")
	// 			if i > -1 {
	// 				payload = payload[i:]
	// 			}

	// 			sugar.Infow("GIT_SSH_REQUEST",
	// 				"session", perms.Extensions["SESSION"],
	// 				"type", req.Type,
	// 				"payload", payload,
	// 				"size", chs.Size(),
	// 				"elapsed", finishTime.Sub(startTime),
	// 				"error", err)
	// 		}()
	// 		err = t(ctx, chs, req, perms, packCmd, repoPath)
	// 		return err
	// 	}
	// }

	// type ChannelWithSize struct {
	// 	ssh.Channel
	// 	size int64
	// }

	// func (ch *ChannelWithSize) Size() int64 {
	// 	return ch.size
	// }

	// func (ch *ChannelWithSize) Write(data []byte) (int, error) {
	// 	written, err := ch.Channel.Write(data)
	// 	ch.size += int64(written)
	// 	return written, err
	// }

	// func exists(filename string) bool {
	// 	_, err := os.Stat(filename)
	// 	return err == nil
	// }

	// func getRemoteAddr(addr net.Addr) string {
	// 	s := addr.String()
	// 	if strings.ContainsRune(s, ':') {
	// 		host, _, _ := net.SplitHostPort(s)
	// 		return host
	// 	}
	// 	return s
}
