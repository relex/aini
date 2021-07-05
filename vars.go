package aini

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// AddVars take a path that contains group_vars and host_vars directories
// and adds these variables to the InventoryData
func (inventory *InventoryData) AddVars(path string) error {
	return inventory.doAddVars(path, false)
}

// AddVarsLowerCased does the same as AddVars, but converts hostnames and groups name to lowercase
// Use this function if you've executed `inventory.HostsToLower` or `inventory.GroupsToLower`
func (inventory *InventoryData) AddVarsLowerCased(path string) error {
	return inventory.doAddVars(path, true)
}

func (inventory *InventoryData) doAddVars(path string, lowercased bool) error {
	_, err := os.Stat(path)
	if err != nil {
		return err
	}
	walk(path, "group_vars", inventory.getGroupsMap(), lowercased)
	walk(path, "host_vars", inventory.getHostsMap(), lowercased)
	return nil
}

type varsGetter interface {
	getVars() map[string]string
}

func (host *Host) getVars() map[string]string {
	return host.Vars
}

func (group *Group) getVars() map[string]string {
	return group.Vars
}

func (inventory InventoryData) getHostsMap() map[string]varsGetter {
	result := make(map[string]varsGetter, len(inventory.Hosts))
	for k, v := range inventory.Hosts {
		result[k] = v
	}
	return result
}

func (inventory InventoryData) getGroupsMap() map[string]varsGetter {
	result := make(map[string]varsGetter, len(inventory.Groups))
	for k, v := range inventory.Groups {
		result[k] = v
	}
	return result
}

func walk(root string, subdir string, m map[string]varsGetter, lowercased bool) error {
	path := filepath.Join(root, subdir)
	_, err := os.Stat(path)
	// If the dir doesn't exist we can just skip it
	if err != nil {
		return nil
	}
	f := getWalkerFn(path, m, lowercased)
	return filepath.WalkDir(path, f)
}

func getWalkerFn(root string, m map[string]varsGetter, lowercased bool) fs.WalkDirFunc {
	var currentVars map[string]string
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if filepath.Dir(path) == root {
			filename := filepath.Base(path)
			ext := filepath.Ext(path)
			itemName := strings.TrimSuffix(filename, ext)
			if lowercased {
				itemName = strings.ToLower(itemName)
			}
			if currentItem, ok := m[itemName]; ok {
				currentVars = currentItem.getVars()
			} else {
				return nil
			}
		}
		if d.IsDir() {
			return nil
		}
		return addVarsFromFile(currentVars, path)
	}
}

func addVarsFromFile(currentVars map[string]string, path string) error {
	if currentVars == nil {
		// Group or Host doesn't exist in the inventory, ignoring
		return nil
	}
	ext := filepath.Ext(path)
	if ext != ".yaml" && ext != ".yml" {
		return nil
	}
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	vars := make(map[string]interface{})
	err = yaml.Unmarshal(f, &vars)
	if err != nil {
		return err
	}
	for k, v := range vars {
		switch v := v.(type) {
		case string:
			currentVars[k] = v
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			currentVars[k] = fmt.Sprint(v)
		case bool:
			currentVars[k] = strconv.FormatBool(v)
		default:
			data, err := json.Marshal(v)
			if err != nil {
				return err
			}
			currentVars[k] = string(data)
		}
	}
	return nil
}
