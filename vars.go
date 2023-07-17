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

// AddVarsLowerCased does the same as AddVars, but converts hostnames and groups name to lowercase.
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
	inventory.reconcileVars()
	return nil
}

type fileVarsGetter interface {
	getFileVars() map[string]string
}

func (host *Host) getFileVars() map[string]string {
	return host.FileVars
}

func (group *Group) getFileVars() map[string]string {
	return group.FileVars
}

func (inventory InventoryData) getHostsMap() map[string]fileVarsGetter {
	result := make(map[string]fileVarsGetter, len(inventory.Hosts))
	for k, v := range inventory.Hosts {
		result[k] = v
	}
	return result
}

func (inventory InventoryData) getGroupsMap() map[string]fileVarsGetter {
	result := make(map[string]fileVarsGetter, len(inventory.Groups))
	for k, v := range inventory.Groups {
		result[k] = v
	}
	return result
}

func walk(root string, subdir string, m map[string]fileVarsGetter, lowercased bool) error {
	path := filepath.Join(root, subdir)
	_, err := os.Stat(path)
	// If the dir doesn't exist we can just skip it
	if err != nil {
		return nil
	}
	f := getWalkerFn(path, m, lowercased)
	return filepath.WalkDir(path, f)
}

func getWalkerFn(root string, m map[string]fileVarsGetter, lowercased bool) fs.WalkDirFunc {
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
				currentVars = currentItem.getFileVars()
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

func (inventory *InventoryData) reconcileVars() {
	/*
		Priority of variables is defined here: https://docs.ansible.com/ansible/latest/user_guide/playbooks_variables.html#understanding-variable-precedence
		Distilled list looks like this:
			1. inventory file group vars
			2. group_vars/*
			3. inventory file host vars
			4. inventory host_vars/*
	*/
	for _, group := range inventory.Groups {
		group.AllInventoryVars = nil
		group.AllFileVars = nil
	}
	for _, group := range inventory.Groups {
		group.Vars = make(map[string]string)
		group.populateInventoryVars()
		group.populateFileVars()
		// At this point we already "populated" all parent's inventory and file vars
		// So it's fine to build Vars right away, without needing the second pass
		group.Vars = copyStringMap(group.AllInventoryVars)
		addValues(group.Vars, group.AllFileVars)
	}
	for _, host := range inventory.Hosts {
		host.Vars = make(map[string]string)
		for _, group := range GroupMapListValues(host.DirectGroups) {
			addValues(host.Vars, group.Vars)
		}
		addValues(host.Vars, host.InventoryVars)
		addValues(host.Vars, host.FileVars)
	}
}

func (group *Group) populateInventoryVars() {
	if group.AllInventoryVars != nil {
		return
	}
	group.AllInventoryVars = make(map[string]string)
	for _, parent := range GroupMapListValues(group.DirectParents) {
		parent.populateInventoryVars()
		addValues(group.AllInventoryVars, parent.AllInventoryVars)
	}
	addValues(group.AllInventoryVars, group.InventoryVars)
}

func (group *Group) populateFileVars() {
	if group.AllFileVars != nil {
		return
	}
	group.AllFileVars = make(map[string]string)
	for _, parent := range GroupMapListValues(group.DirectParents) {
		parent.populateFileVars()
		addValues(group.AllFileVars, parent.AllFileVars)
	}
	addValues(group.AllFileVars, group.FileVars)
}
