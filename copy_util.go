package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func CopyDir(scrDir, dest string) error {
	if entries, err := ioutil.ReadDir(scrDir); err != nil {
		return err
	} else {
		for _, entry := range entries {
			sourcePath := filepath.Join(scrDir, entry.Name())
			destPath := filepath.Join(dest, entry.Name())

			switch entry.Mode() & os.ModeType {
			case os.ModeDir:
				if err := CreateIfNotExists(destPath, entry.Mode()); err != nil {
					return err
				}
				if err := CopyDir(sourcePath, destPath); err != nil {
					return err
				}
			case os.ModeSymlink:
				if err := CopySymLink(sourcePath, destPath); err != nil {
					return err
				}
			default:
				if err := CopyFile(sourcePath, destPath, entry.Mode()); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func CopyFile(srcFile, dstFile string, mode os.FileMode) error {
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
		if err := os.Symlink(link, dest); err != nil && err != os.ErrExist {
			fmt.Printf("Symlink %s to %s failed: %v", link, dest, err)
			return err
		}
	}
	return nil
}
