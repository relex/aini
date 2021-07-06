package aini

// Inventory-related helper methods

// Reconcile ensures inventory basic rules, run after updates
// After initial inventory file processing, only direct relationships are set
// This method sets Children and Parents
func (inventory *InventoryData) Reconcile() {
	allGroup := inventory.getOrCreateGroup("all")
	allGroup.Hosts = inventory.Hosts
	allGroup.Children = inventory.Groups

	for _, host := range inventory.Hosts {
		for _, group := range host.directGroups {
			group.Hosts[host.Name] = host
			host.Groups[group.Name] = group
			group.directParents[allGroup.Name] = allGroup
			for _, ancestor := range group.getAncestors() {
				group.Parents[ancestor.Name] = ancestor
				ancestor.Children[group.Name] = group
				ancestor.Hosts[host.Name] = host
				host.Groups[ancestor.Name] = ancestor
			}
		}
	}
	inventory.reconcileVars()
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

	for queue := []*Group{group}; ; {
		group := queue[0]
		parentList := GroupMapListValues(group.directParents)
		result = append(result, parentList...)
		copy(queue, queue[1:])
		queue = queue[:len(queue)-1]
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
