//go:build !windows
// +build !windows

package plugin

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"sync"
)

var ManagementPlugin = managementPlugin{mp: make(map[string]*plugin.Plugin)}

type managementPlugin struct {
	mp map[string]*plugin.Plugin
	sync.Mutex
}

func (m *managementPlugin) SetPlugin(key string, p *plugin.Plugin) {
	m.Lock()
	defer m.Unlock()
	m.mp[key] = p
}

func (m *managementPlugin) GetPlugin(key string) (p *plugin.Plugin, ok bool) {
	m.Lock()
	defer m.Unlock()
	p, ok = m.mp[key]
	return
}

// LoadPlugin Load the plugin and pass in the path
func LoadPlugin(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		fileSlice, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}
		for _, ff := range fileSlice {
			if !ff.IsDir() && filepath.Ext(ff.Name()) == ".so" {
				if err = loadPlugin(path, ff); err != nil {
					return err
				}
			} else if ff.IsDir() {
				_ = LoadPlugin(filepath.Join(path, ff.Name()))
			}
		}
		return nil
	} else {
		return loadPlugin(path, fileInfo)
	}
}

func loadPlugin(path string, f fs.FileInfo) error {
	if filepath.Ext(f.Name()) == ".so" {
		fPath := filepath.Join(path, f.Name())
		// Loading plugins
		p, err := plugin.Open(fPath)
		if err != nil {
			fmt.Println("loadPlugin err ", err)
			return err
		}
		// Determine if the protocol is met
		// To meet the requirements of OnlyFuncName && Implement Plugin Interface
		if v, err := p.Lookup(OnlyFuncName); err != nil {
			fmt.Println("loadPlugin err ", err)
			return err
		} else if _, ok := v.(Plugin); !ok {
			fmt.Println("loadPlugin err ", fmt.Sprintf("path:%s no implementation of the %s interface", filepath.Base(fPath), OnlyFuncName))
			return errors.New("Not implementing the specified interface")
		} else {
			// todo
			fmt.Println("todo...")
		}
		fmt.Println("loadPlugin add ", filepath.Base(fPath))
		ManagementPlugin.SetPlugin(filepath.Base(fPath), p)
	}
	return nil
}
