package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/relex/aini"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
)

func main() {
	if len(os.Args) > 3 || len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: ainidump inventory_file [host_or_group_patterns]")
		os.Exit(1)
	}

	inventoryPath, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve inventory file path %s: %v\n", os.Args[1], err)
		os.Exit(2)
	}

	inventory, err := aini.ParseFile(inventoryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse inventory file %s: %v\n", inventoryPath, err)
		os.Exit(3)
	}

	inventory.HostsToLower()
	inventory.GroupsToLower()

	inventoryDir := filepath.Dir(inventoryPath)
	if err := inventory.AddVarsLowerCased(inventoryDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load inventory variables %s: %v\n", inventoryDir, err)
		os.Exit(4)
	}

	if len(os.Args) == 2 {
		result := exportResult(inventory.Hosts, inventory.Groups)
		j, err := json.MarshalIndent(result, "", "    ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(j))
		return
	}

	patterns := os.Args[2]

	matchedHostsMap, err := inventory.MatchHostsByPatterns(patterns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to match hosts with patterns %s: %v\n", patterns, err)
		os.Exit(5)
	}
	j, err := json.MarshalIndent(lo.MapEntries(matchedHostsMap, func(name string, host *aini.Host) (string, ResultHost) {
		return host.Name, ResultHost{
			Name:   host.Name,
			Groups: getGroupNames(host.ListGroupsOrdered()),
			Vars:   host.Vars,
		}
	}), "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(j))
}

type ResultHost struct {
	Name   string
	Groups []string
	Vars   map[string]string
}
type ResultGroup struct {
	Name        string
	Parents     []string
	Descendants []string
	Hosts       []string
	Vars        map[string]string
}

func exportResult(hostMap map[string]*aini.Host, groupMap map[string]*aini.Group) any {
	type Result struct {
		Hosts  []ResultHost
		Groups []ResultGroup
	}
	result := &Result{
		Hosts:  make([]ResultHost, 0, len(hostMap)),
		Groups: make([]ResultGroup, 0, len(groupMap)),
	}

	orderedHosts := maps.Values(hostMap)
	slices.SortStableFunc(orderedHosts, func(a, b *aini.Host) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, host := range orderedHosts {
		result.Hosts = append(result.Hosts, ResultHost{
			Name:   host.Name,
			Groups: getGroupNames(host.ListGroupsOrdered()),
			Vars:   host.Vars,
		})
	}

	orderedGroups := maps.Values(groupMap)
	slices.SortStableFunc(orderedGroups, func(a, b *aini.Group) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, group := range orderedGroups {
		orderedDescendantNames := getGroupNames(maps.Values(group.Children))
		sort.Strings(orderedDescendantNames)

		orderedHostNames := getHostNames(maps.Values(group.Hosts))
		sort.Strings(orderedHostNames)

		result.Groups = append(result.Groups, ResultGroup{
			Name:        group.Name,
			Parents:     getGroupNames(group.ListParentGroupsOrdered()),
			Descendants: orderedDescendantNames,
			Hosts:       orderedHostNames,
			Vars:        group.Vars,
		})
	}

	return &result
}

func getGroupNames(groups []*aini.Group) []string {
	groupNames := make([]string, 0, len(groups))
	for _, grp := range groups {
		groupNames = append(groupNames, grp.Name)
	}
	return groupNames
}

func getHostNames(hosts []*aini.Host) []string {
	hostNames := make([]string, 0, len(hosts))
	for _, hst := range hosts {
		hostNames = append(hostNames, hst.Name)
	}
	return hostNames
}
