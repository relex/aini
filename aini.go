package aini

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"path"
	"strings"
)

// InventoryData contains parsed inventory representation
// Note: Groups and Hosts fields contain all the groups and hosts, not only top-level
type InventoryData struct {
	Groups map[string]*Group
	Hosts  map[string]*Host
}

// Group represents ansible group
type Group struct {
	Name     string
	Vars     map[string]string
	Hosts    map[string]*Host
	Children map[string]*Group
	Parents  map[string]*Group
}

// Host represents ansible host
type Host struct {
	Name   string
	Port   int
	Vars   map[string]string
	Groups map[string]*Group
}

// ParseFile parses Inventory represented as a file
func ParseFile(f string) (*InventoryData, error) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return &InventoryData{}, err
	}

	return Parse(bytes.NewReader(bs))
}

// ParseString parses Inventory represented as a string
func ParseString(input string) (*InventoryData, error) {
	return Parse(strings.NewReader(input))
}

// Parse using some Reader
func Parse(r io.Reader) (*InventoryData, error) {
	input := bufio.NewReader(r)
	inventory := &InventoryData{}
	err := inventory.parse(input)
	if err != nil {
		return inventory, err
	}
	inventory.Reconcile()
	return inventory, nil
}

// Match looks for a hosts that match the pattern
func (inventory *InventoryData) Match(m string) []*Host {
	matchedHosts := make([]*Host, 0)
	for _, host := range inventory.Hosts {
		if m, err := path.Match(m, host.Name); err == nil && m {
			matchedHosts = append(matchedHosts, host)
		}
	}
	return matchedHosts
}
