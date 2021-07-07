package aini

// Inventory-related helper methods

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
			for _, ancestor := range group.getAncestors() {
				inventory.Groups[ancestor.Name] = ancestor
			}
		}
	}

	// Calculate intergroup relationships
	for _, group := range inventory.Groups {
		group.directParents[allGroup.Name] = allGroup
		for _, ancestor := range group.getAncestors() {
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

// getAncestors returns all Ancestors of a given group in level order
func (group *Group) getAncestors() []*Group {
	result := make([]*Group, 0)
	if len(group.directParents) == 0 {
		return result
	}
	visited := map[string]struct{}{group.Name: {}}

	for queue := GroupMapListValues(group.directParents); ; {
		group := queue[0]
		copy(queue, queue[1:])
		queue = queue[:len(queue)-1]
		if _, ok := visited[group.Name]; ok {
			if len(queue) == 0 {
				return result
			}
			continue
		}
		visited[group.Name] = struct{}{}
		parentList := GroupMapListValues(group.directParents)
		result = append(result, group)
		queue = append(queue, parentList...)

		if len(queue) == 0 {
			return result
		}
	}
}

// addValues fills `to` map with values from `from` map
func addValues(to map[string]string, from map[string]string) {
	for k, v := range from {
		to[k] = v
	}
}

// copyStringMap creates a non-deep copy of the map
func copyStringMap(from map[string]string) map[string]string {
	result := make(map[string]string, len(from))
	addValues(result, from)
	return result
}
