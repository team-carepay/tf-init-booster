package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	gitcrypt "github.com/jbuchbinder/go-git-crypt"
)

var (
	gitCryptHeader = []byte{0, 'G', 'I', 'T', 'C', 'R', 'Y', 'P', 'T'}
)

const (
	aesEncryptorNonceLen = 12
)

type Copier struct {
	gc  *gitcrypt.GitCrypt
	key gitcrypt.Key
}

func NewCopier() (*Copier, error) {
	keyPath := os.Getenv("GIT_CRYPT_KEY")
	if keyPath != "" {
		gc := gitcrypt.GitCrypt{}
		key, err := gc.KeyFromFile(keyPath)
		key.Version = 0 // not sure why this is needed
		if err != nil {
			return nil, err
		}
		return &Copier{
			gc:  &gc,
			key: key,
		}, nil
	}
	return &Copier{}, nil
}

func (copier *Copier) CopyDir(scrDir, dest string) error {
	if entries, err := os.ReadDir(scrDir); err != nil {
		return err
	} else {
		for _, entry := range entries {
			sourcePath := filepath.Join(scrDir, entry.Name())
			destPath := filepath.Join(dest, entry.Name())

			info, _ := entry.Info()
			switch entry.Type() & os.ModeType {
			case os.ModeDir:
				if err := CreateIfNotExists(destPath, info.Mode()); err != nil {
					return err
				}
				if err := copier.CopyDir(sourcePath, destPath); err != nil {
					return err
				}
			case os.ModeSymlink:
				if err := CopySymLink(sourcePath, destPath); err != nil {
					return err
				}
			default:
				if err := copier.CopyFile(sourcePath, destPath, info.Mode()); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func (copier *Copier) CopyFile(srcFile, dstFile string, mode os.FileMode) error {
	if out, err := os.OpenFile(dstFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode); err != nil {
		return err
	} else {
		defer out.Close()
		if in, err := os.Open(srcFile); err != nil {
			return err
		} else {
			defer in.Close()
			header := make([]byte, 10+aesEncryptorNonceLen)
			n, err := in.Read(header)
			if err != nil && err != io.EOF {
				return fmt.Errorf("failed to read header from file: '%s', n is %d, error: '%s'", srcFile, n, err.Error())
			}
			if copier.gc != nil && bytes.Equal(header[0:9], gitCryptHeader) {
				err = copier.gc.DecryptStream(copier.key, header, in, out)
				if err != nil {
					fmt.Printf("failed to decrypt file: '%s', error: '%s'", srcFile, err.Error())
				}
			} else if n > 0 {
				_, err = out.Write(header[0:n]) // write the bytes we read earlier to check for the header
				if err != nil {
					return fmt.Errorf("failed to write header to file: '%s', error: '%s'", dstFile, err.Error())
				}
				if _, err = io.Copy(out, in); err != nil {
					return fmt.Errorf("failed to copy file: '%s', error: '%s'", srcFile, err.Error())
				}
			}
			return nil
		}
	}
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}
	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}
	return nil
}

func CopySymLink(source, dest string) error {
	if link, err := os.Readlink(source); err != nil {
		return err
	} else {
		_ = os.Remove(dest)
		if err := os.Symlink(link, dest); err != nil && !os.IsExist(err) {
			return err
		}
	}
	return nil
}
