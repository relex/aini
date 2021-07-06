# aini

Go library for Parsing Ansible inventory files.  
We are trying to follow the logic of Ansible parser as close as possible.

Documentation on ansible inventory files can be found here:  
https://docs.ansible.com/ansible/latest/user_guide/intro_inventory.html

## Supported features:
- [X] Variables
- [X] Host patterns
- [X] Nested groups
- [X] Load variables from `group_vars` and `host_vars`

## Public API
```godoc
package aini // import "github.com/relex/aini"


TYPES

type Group struct {
        Name     string
        Vars     map[string]string
        Hosts    map[string]*Host
        Children map[string]*Group
        Parents  map[string]*Group

        // Has unexported fields.
}
    Group represents ansible group

func GroupMapListValues(mymap map[string]*Group) []*Group
    GroupMapListValues transforms map of Groups into Group list in lexical order

type Host struct {
        Name   string
        Port   int
        Vars   map[string]string
        Groups map[string]*Group

        // Has unexported fields.
}
    Host represents ansible host

func HostMapListValues(mymap map[string]*Host) []*Host
    HostMapListValues transforms map of Hosts into Host list in lexical order

type InventoryData struct {
        Groups map[string]*Group
        Hosts  map[string]*Host
}
    InventoryData contains parsed inventory representation Note: Groups and
    Hosts fields contain all the groups and hosts, not only top-level

func Parse(r io.Reader) (*InventoryData, error)
    Parse using some Reader

func ParseFile(f string) (*InventoryData, error)
    ParseFile parses Inventory represented as a file

func ParseString(input string) (*InventoryData, error)
    ParseString parses Inventory represented as a string

func (inventory *InventoryData) AddVars(path string) error
    AddVars take a path that contains group_vars and host_vars directories and
    adds these variables to the InventoryData

func (inventory *InventoryData) AddVarsLowerCased(path string) error
    AddVarsLowerCased does the same as AddVars, but converts hostnames and
    groups name to lowercase Use this function if you've executed
    `inventory.HostsToLower` or `inventory.GroupsToLower`

func (inventory *InventoryData) GroupsToLower()
    GroupsToLower transforms all group names to lowercase

func (inventory *InventoryData) HostsToLower()
    HostsToLower transforms all host names to lowercase

func (inventory *InventoryData) Match(m string) []*Host
    Match looks for a hosts that match the pattern

func (inventory *InventoryData) Reconcile()
    Reconcile ensures inventory basic rules, run after updates After initial
    inventory file processing, only direct relationships are set This method
    sets Children and Parents

```

## Usage example
```go
import (
    "strings"
    
    "github.com/relex/aini"
)

func main() {
    // Load from string example
    inventoryReader := strings.NewReader(`
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
