package aini

// Inventory-related helper methods

// Reconcile ensures inventory basic rules, run after updates
func (inventory *InventoryData) Reconcile() {
	allGroup := inventory.getOrCreateGroup("all")
	allGroup.Hosts = inventory.Hosts
	allGroup.Children = inventory.Groups

	for _, host := range inventory.Hosts {
		for _, group := range host.Groups {
			host.setVarsIfNotExist(group.Vars)
			for _, ancestor := range group.getAncestors() {
				ancestor.Hosts[host.Name] = host
				ancestor.Children[group.Name] = group
				host.Groups[ancestor.Name] = ancestor
				for k, v := range ancestor.Vars {
					if _, ok := host.Vars[k]; !ok {
						host.Vars[k] = v
					}
					if _, ok := group.Vars[k]; !ok {
						group.Vars[k] = v
					}
				}
			}
		}
		host.setVarsIfNotExist(allGroup.Vars)
		host.Groups["all"] = allGroup
	}
	inventory.Groups["all"] = allGroup
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
	}
	inventory.Hosts[hostName] = h
	return h
}

// getAncestors returns all Ancestors of a given group in level order
func (group *Group) getAncestors() []*Group {
	result := make([]*Group, 0)

	for queue := []*Group{group}; ; {
		group := queue[0]
		parentList := GroupMapListValues(group.Parents)
		result = append(result, parentList...)
		copy(queue, queue[1:])
		queue = queue[:len(queue)-1]
		queue = append(queue, parentList...)

		if len(queue) == 0 {
			return result
		}
	}
}

// setVarsIfNotExist sets Var for host if it doesn't have it already
func (host *Host) setVarsIfNotExist(vars map[string]string) {
	for k, v := range vars {
		if _, ok := host.Vars[k]; !ok {
			host.Vars[k] = v
		}
	}
}

func addValuesFromMap(m1 map[string]string, m2 map[string]string) {
	for k, v := range m2 {
		if m1[k] == "" {
			m1[k] = v
		}
	}
}
