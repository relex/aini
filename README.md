# aini

Go library for Parsing Ansible inventory files.  
We are trying to follow the logic of Ansible parser as close as possible.

Documentation on ansible inventory files can be found here:  
https://docs.ansible.com/ansible/latest/user_guide/intro_inventory.html

## Supported features:
- [X] Variables
- [X] Host patterns
- [X] Nested groups

## Public API
```go
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

// ParseFile parses the file
func ParseFile(f string) (*InventoryData, error) {
    ...
}

// Parse using some Reader
func Parse(r io.Reader) (*InventoryData, error) {
    ...
}

// Match looks for a hosts that match the pattern
func (inventory *InventoryData) Match(m string) []*Host {
    ..
}

// Reconcile ensures inventory basic rules, run after updates
func (inventory *InventoryData) Reconcile() {
    ...
}
```

## Usage example
```go
import (
    "strings"
    
    "github.com/relex/aini/aini"
)

func main() {
    // Load from string example
    inventoryReader := strings.NewReader((`
	host1:2221
	[web]
	host2 ansible_ssh_user=root
    `)
    var inventory InventoryData = aini.Parse(inventoryReader)

    // Querying hosts
    _ = inventory.Hosts["host1"].Name == "host1"  // true
    _ = inventory.Hosts["host1"].Port == 2221     // true
    _ = inventory.Hosts["host2"].Name == "host2"] // true
    _ = inventory.Hosts["host2"].Post == 22]      // true
    
    _ = len(inventory.Hosts["host1"].Groups) == 2 // all, ungrouped
    _ = len(inventory.Hosts["host2"].Groups) == 2 // all, web
    
    _ = len(inventory.Match("host*")) == 2        // host1, host2

    _ = // Querying groups
    _ = inventory.Groups["web"].Hosts[0].Name == "host2" // true
    _ = len(inventory.Groups["all"].Hosts) == 2          // true
}
```
