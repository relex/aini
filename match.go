package aini

import "path"

// MatchHosts looks for hosts that match the pattern
func (inventory *InventoryData) MatchHosts(pattern string) (map[string]*Host, error) {
	return MatchHosts(inventory.Hosts, pattern)
}

// MatchHosts looks for hosts that match the pattern
func (group *Group) MatchHosts(pattern string) (map[string]*Host, error) {
	return MatchHosts(group.Hosts, pattern)
}

// MatchHosts looks for hosts that match the pattern
func MatchHosts(hosts map[string]*Host, pattern string) (map[string]*Host, error) {
	matchedHosts := make(map[string]*Host)
	for _, host := range hosts {
		m, err := path.Match(pattern, host.Name)
		if err != nil {
			return nil, err
		}
		if m {
			matchedHosts[host.Name] = host
		}
	}
	return matchedHosts, nil
}

// MatchGroups looks for groups that match the pattern
func (inventory *InventoryData) MatchGroups(pattern string) (map[string]*Group, error) {
	return MatchGroups(inventory.Groups, pattern)
}

// MatchGroups looks for groups that match the pattern
func (host *Host) MatchGroups(pattern string) (map[string]*Group, error) {
	return MatchGroups(host.Groups, pattern)
}

// MatchGroups looks for groups that match the pattern
func MatchGroups(groups map[string]*Group, pattern string) (map[string]*Group, error) {
	matchedGroups := make(map[string]*Group)
	for _, group := range groups {
		m, err := path.Match(pattern, group.Name)
		if err != nil {
			return nil, err
		}
		if m {
			matchedGroups[group.Name] = group
		}
	}
	return matchedGroups, nil
}

// MatchVars looks for vars that match the pattern
func (group *Group) MatchVars(pattern string) (map[string]string, error) {
	return MatchVars(group.Vars, pattern)
}

// MatchVars looks for vars that match the pattern
func (host *Host) MatchVars(pattern string) (map[string]string, error) {
	return MatchVars(host.Vars, pattern)
}

// MatchVars looks for vars that match the pattern
func MatchVars(vars map[string]string, pattern string) (map[string]string, error) {
	matchedVars := make(map[string]string)
	for k, v := range vars {
		m, err := path.Match(pattern, v)
		if err != nil {
			return nil, err
		}
		if m {
			matchedVars[k] = v
		}
	}
	return matchedVars, nil
}
