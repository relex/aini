package aini

import "sync"

// Inventory-related helper methods

var mapPool = sync.Pool{
    New: func() interface{} {
        return make(map[string]string)
    },
}

// Reconcile ensures inventory basic rules, run after updates.
// After initial inventory file processing, only direct relationships are set.
//
// This method:
//   * (re)sets Children and Parents for hosts and groups
//   * ensures that mandatory groups exist
//   * calculates variables for hosts and groups
func (inventory *InventoryData) Reconcile() {
	// Clear all computed data
	for _, host := range inventory.Hosts {
		host.clearData()
	}
	// a group can be empty (with no hosts in it), so the previous method will not clean it
	// on the other hand, a group could have been attached to a host by a user, but not added to the inventory.Groups map
	// so it's safer just to clean everything
	for _, group := range inventory.Groups {
		group.clearData(make(map[string]struct{}, len(inventory.Groups)))
	}

	allGroup := inventory.getOrCreateGroup("all")
	ungroupedGroup := inventory.getOrCreateGroup("ungrouped")
	ungroupedGroup.directParents[allGroup.Name] = allGroup

	// First, ensure that inventory.Groups contains all the groups
	for _, host := range inventory.Hosts {
		for _, group := range host.directGroups {
			inventory.Groups[group.Name] = group
			for _, ancestor := range group.ListParentGroupsOrdered() {
				inventory.Groups[ancestor.Name] = ancestor
			}
		}
	}

	// Calculate intergroup relationships
	for _, group := range inventory.Groups {
		group.directParents[allGroup.Name] = allGroup
		for _, ancestor := range group.ListParentGroupsOrdered() {
			group.Parents[ancestor.Name] = ancestor
			ancestor.Children[group.Name] = group
		}
	}

	// Now set hosts for groups and groups for hosts
	for _, host := range inventory.Hosts {
		host.Groups[allGroup.Name] = allGroup
		for _, group := range host.directGroups {
			group.Hosts[host.Name] = host
			host.Groups[group.Name] = group
			for _, parent := range group.Parents {
				group.Parents[parent.Name] = parent
				parent.Children[group.Name] = group
				parent.Hosts[host.Name] = host
				host.Groups[parent.Name] = parent
			}
		}
	}
	inventory.reconcileVars()
}

func (host *Host) clearData() {
	host.Groups = make(map[string]*Group)
	host.Vars = make(map[string]string)
	for _, group := range host.directGroups {
		group.clearData(make(map[string]struct{}, len(host.Groups)))
	}
}

func (group *Group) clearData(visited map[string]struct{}) {
	if _, ok := visited[group.Name]; ok {
		return
	}
	group.Hosts = make(map[string]*Host)
	group.Parents = make(map[string]*Group)
	group.Children = make(map[string]*Group)
	group.Vars = make(map[string]string)
	group.allInventoryVars = nil
	group.allFileVars = nil
	visited[group.Name] = struct{}{}
	for _, parent := range group.directParents {
		parent.clearData(visited)
	}
}

// getOrCreateGroup return group from inventory if exists or creates empty Group with given name
func (inventory *InventoryData) getOrCreateGroup(groupName string) *Group {
	if group, ok := inventory.Groups[groupName]; ok {
		return group
	}
	g := &Group{
		Name:     groupName,
		Hosts:    make(map[string]*Host),
		Vars:     make(map[string]string),
		Children: make(map[string]*Group),
		Parents:  make(map[string]*Group),

		directParents: make(map[string]*Group),
		inventoryVars: make(map[string]string),
		fileVars:      make(map[string]string),
	}
	inventory.Groups[groupName] = g
	return g
}

// getOrCreateHost return host from inventory if exists or creates empty Host with given name
func (inventory *InventoryData) getOrCreateHost(hostName string) *Host {
	if host, ok := inventory.Hosts[hostName]; ok {
		return host
	}
	h := &Host{
		Name:   hostName,
		Port:   22,
		Groups: make(map[string]*Group),
		Vars:   make(map[string]string),

		directGroups:  make(map[string]*Group),
		inventoryVars: make(map[string]string),
		fileVars:      make(map[string]string),
	}
	inventory.Hosts[hostName] = h
	return h
}

// addValues fills `to` map with values from `from` map
func addValues(to map[string]string, from map[string]string) {
    // Get a map from the pool
    m := mapPool.Get().(map[string]string)

    // Copy values from `from` to `m`
    for k, v := range from {
        m[k] = v	
    }

    // Copy values from `m` to `to`
    for k, v := range m {
        to[k] = v
    }

    // Reset the map and put it back in the pool
    for k := range m {
        delete(m, k)
    }
    mapPool.Put(m)
}


// copyStringMap creates a non-deep copy of the map
func copyStringMap(from map[string]string) map[string]string {
	result := make(map[string]string, len(from))
	addValues(result, from)
	return result
}
