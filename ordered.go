package aini

import (
	"path"
)

// MatchGroupsOrdered looks for groups that match the pattern
// The result is a sorted array, where lower indexes corespond to more specific groups
func (host *Host) MatchGroupsOrdered(pattern string) ([]*Group, error) {
	matchedGroups := make([]*Group, 0)
	groups := host.ListGroupsOrdered()

	for _, group := range groups {
		m, err := path.Match(pattern, group.Name)
		if err != nil {
			return nil, err
		}
		if m {
			matchedGroups = append(matchedGroups, group)
		}
	}

	return matchedGroups, nil
}

// MatchGroupsOrdered looks for groups that match the pattern
// The result is a sorted array, where lower indexes corespond to more specific groups
func (group *Group) MatchGroupsOrdered(pattern string) ([]*Group, error) {
	matchedGroups := make([]*Group, 0)
	groups := group.ListParentGroupsOrdered()

	for _, group := range groups {
		m, err := path.Match(pattern, group.Name)
		if err != nil {
			return nil, err
		}
		if m {
			matchedGroups = append(matchedGroups, group)
		}
	}

	return matchedGroups, nil
}

// ListGroupsOrdered returns all ancestor groups of a given host in level order
func (host *Host) ListGroupsOrdered() []*Group {
	return listAncestorsOrdered(host.directGroups, nil, true)
}

// ListParentGroupsOrdered returns all ancestor groups of a given group in level order
func (group *Group) ListParentGroupsOrdered() []*Group {
	visited := map[string]struct{}{group.Name: {}}
	return listAncestorsOrdered(group.directParents, visited, group.Name != "all")
}

// listAncestorsOrdered returns all ancestor groups of a given group map in level order
func listAncestorsOrdered(groups map[string]*Group, visited map[string]struct{}, appendAll bool) []*Group {
	result := make([]*Group, 0)
	if visited == nil {
		visited = map[string]struct{}{}
	}
	var allGroup *Group
	for queue := GroupMapListValues(groups); len(queue) > 0; func() {
		copy(queue, queue[1:])
		queue = queue[:len(queue)-1]
	}() {
		group := queue[0]
		// The all group should always be the last one
		if group.Name == "all" {
			allGroup = group
			continue
		}
		if _, ok := visited[group.Name]; ok {
			continue
		}
		visited[group.Name] = struct{}{}
		parentList := GroupMapListValues(group.directParents)
		result = append(result, group)
		queue = append(queue, parentList...)
	}
	if allGroup != nil && appendAll {
		result = append(result, allGroup)
	}
	return result
}
