package aini

import (
	"fmt"
	"path"
	"strings"

	"golang.org/x/exp/maps"
)

// MatchHostsByPatterns looks for all hosts that match the Ansible host patterns as described in https://docs.ansible.com/ansible/latest/inventory_guide/intro_patterns.html
//
// e.g. "webservers:gateways:myhost.domain:!atlanta"
func (inventory *InventoryData) MatchHostsByPatterns(patterns string) (map[string]*Host, error) {
	patternList := strings.Split(patterns, ":")

	matchedHosts := make(map[string]*Host)
	for _, host := range inventory.Hosts {
		matched, err := host.MatchPatterns(patternList)
		if err != nil {
			return matchedHosts, err
		}
		if matched {
			matchedHosts[host.Name] = host
		}
	}
	return matchedHosts, nil
}

// MatchPatterns checks whether the given host matches the list of Ansible host patterns.
//
// e.g. [webservers, gateways, myhost.domain, !atlanta]
func (host *Host) MatchPatterns(patterns []string) (bool, error) {
	allNames := make([]string, 0, 1+len(host.Groups))
	allNames = append(allNames, host.Name)
	allNames = append(allNames, maps.Keys(host.Groups)...)
	return MatchNamesByPatterns(allNames, patterns)
}

// MatchNamesByPatterns checks whether the give hostname and group names match list of Ansible host patterns.
//
// e.g. [webservers, gateways, myhost.domain, !atlanta]
func MatchNamesByPatterns(allNames []string, patterns []string) (bool, error) {
	numPositiveMatch := 0

	for index, pattern := range patterns {
		switch {
		case pattern == "":
			if index == 0 {
				return false, nil
			}
			continue
		case pattern == "all" || pattern == "*":
			numPositiveMatch++
		case pattern[0] == '!':
			if index == 0 {
				return false, fmt.Errorf("exclusion pattern \"%s\" cannot be the first pattern", pattern)
			}
			any, err := matchAnyName(pattern[1:], allNames)
			if err != nil {
				return false, err
			}
			if any {
				return false, err
			}
		case pattern[0] == '&':
			if index == 0 {
				return false, fmt.Errorf("intersection pattern \"%s\" cannot be the first pattern", pattern)
			}
			any, err := matchAnyName(pattern[1:], allNames)
			if err != nil {
				return false, err
			}
			if !any {
				return false, err
			}
		default:
			any, err := matchAnyName(pattern, allNames)
			if err != nil {
				return false, err
			}
			if any {
				numPositiveMatch++
			}
		}
	}
	return numPositiveMatch > 0, nil
}

func matchAnyName(pattern string, allNames []string) (bool, error) {
	for _, name := range allNames {
		matched, err := path.Match(pattern, name)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

// MatchHosts looks for hosts whose hostnames match the pattern. Group memberships are not considered.
func (inventory *InventoryData) MatchHosts(pattern string) (map[string]*Host, error) {
	return MatchHosts(inventory.Hosts, pattern)
}

// MatchHosts looks for hosts whose hostnames match the pattern. Group memberships are not considered.
func (group *Group) MatchHosts(pattern string) (map[string]*Host, error) {
	return MatchHosts(group.Hosts, pattern)
}

// MatchHosts looks for hosts whose hostnames match the pattern. Group memberships are not considered.
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
