package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type ModuleRef struct {
	Name      string
	Host      string
	Path      string
	Branch    string
	SubModule string
	Dir       string
}

func NewModuleRef(name, host, path, branch, submodule, dir string) *ModuleRef {
	return &ModuleRef{
		Name:      name,
		Host:      host,
		Path:      path,
		Branch:    branch,
		SubModule: submodule,
		Dir:       dir,
	}
}

func (m *ModuleRef) ToModule() *Module {
	moduleDir := m.Dir
	if strings.HasPrefix(m.SubModule, "//") {
		moduleDir = filepath.Join(m.Dir, m.SubModule[2:])
	}

	return &Module{
		Key:    m.Name,
		Source: fmt.Sprintf("git@%s:%s.git%s%s", m.Host, m.Path, m.SubModule, m.Branch),
		Dir:    moduleDir,
	}
}

func WriteModules(modules []*ModuleRef, file string) error {
	var moduleArray []*Module
	for _, v := range modules {
		moduleArray = append(moduleArray, v.ToModule())
	}
	if modulesJson, err := json.Marshal(Modules{Modules: moduleArray}); err != nil {
		return err
	} else {
		if err := ioutil.WriteFile(file, modulesJson, os.ModePerm); err != nil {
			return err
		} else {
			return nil
		}
	}
}

type Modules struct {
	Modules []*Module `json:"Modules"`
}

type Module struct {
	Key    string `json:"Key"`
	Source string `json:"Source"`
	Dir    string `json:"Dir"`
}
