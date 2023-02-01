package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	moduleutil "github.com/team-carepay/tf-init-booster/internal/moduleutil"
	repo "github.com/team-carepay/tf-init-booster/internal/repository"
)

func main() {
	auth := repo.GetAuth{}

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	modules, err := moduleutil.ScanModules()
	if err != nil {
		log.Fatal(err)
	}
	if len(modules) > 0 {
		if len(os.Args) == 2 {
			for _, m := range modules {
				m.Dir = os.Args[1]
			}
		} else {
			if err := moduleutil.CopyModules(modules, filepath.Join(usr.HomeDir, ".terraform.d/repositories"), auth.Get); err != nil {
				log.Fatal(err)
			}
		}
		if err := moduleutil.WriteModules(modules, ".terraform/modules/modules.json"); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("No modules found, skipping booster")
	}
}
