package aini

// Inventory-related helper methods

// Reconcile ensures inventory basic rules, run after updates
func (inventory *InventoryData) Reconcile() {
	allGroup := inventory.getGroup("all")
	allGroup.Hosts = inventory.Hosts
	allGroup.Children = inventory.Groups

	for _, host := range inventory.Hosts {
		for _, group := range host.Groups {
			ancestors := group.getAncestors()
			host.setVarsIfNotExist(group.Vars)
			for _, ancestor := range ancestors {
				ancestor.Hosts[host.Name] = host
				ancestor.Children[group.Name] = group
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

// getGroup return group from inventory if exists or creates empty Group with given name
func (inventory *InventoryData) getGroup(groupName string) *Group {
	if group, ok := inventory.Groups[groupName]; ok {
		return group
	}
	return &Group{
		Name:     groupName,
		Hosts:    make(map[string]*Host, 0),
		Vars:     make(map[string]string, 0),
		Children: make(map[string]*Group, 0),
		Parents:  make(map[string]*Group, 0),
	}

}

// getHost return host from inventory if exists or creates empty Host with given name
func (inventory *InventoryData) getHost(hostName string) *Host {
	if host, ok := inventory.Hosts[hostName]; ok {
		return host
	}
	return &Host{
		Name:   hostName,
		Port:   22,
		Groups: make(map[string]*Group, 0),
		Vars:   make(map[string]string, 0),
	}

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
