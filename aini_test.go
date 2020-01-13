package aini

import (
	"fmt"
	"strings"
	"testing"
)

func parseString(input string) InventoryData {
	testInput := strings.NewReader(input)
	i, _ := Parse(testInput)
	return *i
}

func (inventory *InventoryData) assertGroupExists(t *testing.T, group string) {
	if inventory.Groups[group] == nil {
		t.Errorf("Cannot find group \"%s\" in %v", group, inventory.Groups)
	}
}

func (host *Host) assertGroupExists(t *testing.T, group string) {
	if host.Groups[group] == nil {
		t.Errorf("Cannot find group \"%s\" in %v", group, host.Groups)
	}
}

func (host *Host) assertVar(t *testing.T, name string, value string) {
	if host.Vars[name] != value {
		t.Errorf("Host %s doesn't have expected variable %s. Expected value: %s, Actual value: %s", host.Name, name, value, host.Vars[name])
	}
}

func (group *Group) assertChildGroupExists(t *testing.T, child string) {
	if group.Children[child] == nil {
		t.Errorf("Cannot find child group \"%s\" in %v", child, group.Parents)
	}
}
func (group *Group) assertParentGroupExists(t *testing.T, parent string) {
	if group.Parents[parent] == nil {
		t.Errorf("Cannot find child group \"%s\" in %v", parent, group.Parents)
	}
}

func (inventory *InventoryData) assertHostExists(t *testing.T, host string) {
	if inventory.Hosts[host] == nil {
		t.Errorf("Cannot find host \"%s\" in %v", host, inventory.Hosts)
	}
}

func (group *Group) assertHostExists(t *testing.T, host string) {
	if group.Hosts[host] == nil {
		t.Errorf("Cannot find host \"%s\" in %v", host, group.Hosts)
	}
}

func assert(t *testing.T, cond bool, msg string) {
	if !cond {
		t.Error(msg)
	}
}

func TestBelongToBasicGroups(t *testing.T) {
	v := parseString(`
	host1:2221 # Comments
	[web]      # should
	host2      # be
	           # ignored
	`)

	assert(t, len(v.Hosts) == 2, "Exactly two hosts expected")
	assert(t, len(v.Groups) == 3, fmt.Sprintf("Expected three groups \"web\", \"all\" and \"ungrouped\", got: %v", v.Groups))

	v.assertGroupExists(t, "web")
	v.assertGroupExists(t, "all")
	v.assertGroupExists(t, "ungrouped")

	v.assertHostExists(t, "host1")
	assert(t, len(v.Hosts["host1"].Groups) == 2, "Host1 must belong to two groups: ungrouped and all")
	assert(t, v.Hosts["host1"].Groups["all"] != nil, "Host1 must belong to all group")
	assert(t, v.Hosts["host1"].Groups["ungrouped"] != nil, "Host1 must belong to ungrouped group")

	v.assertHostExists(t, "host2")
	assert(t, len(v.Hosts["host2"].Groups) == 2, "Host2 must belong to two groups: web and all")
	assert(t, v.Hosts["host2"].Groups["all"] != nil, "Host2 must belong to all group")
	assert(t, v.Hosts["host2"].Groups["web"] != nil, "Host1 must belong to web group")

	assert(t, len(v.Groups["all"].Hosts) == 2, "Group all must contain two hosts")
	v.Groups["all"].assertHostExists(t, "host1")
	v.Groups["all"].assertHostExists(t, "host2")

	assert(t, len(v.Groups["web"].Hosts) == 1, "Group web must contain one host")
	v.Groups["web"].assertHostExists(t, "host2")

	assert(t, len(v.Groups["ungrouped"].Hosts) == 1, "Group ungrouped must contain one host")
	v.Groups["ungrouped"].assertHostExists(t, "host1")

	assert(t, v.Hosts["host1"].Port == 2221, "Host1 ports doesn't match")
	assert(t, v.Hosts["host2"].Port == 22, "Host2 ports doesn't match")
}

func TestGroupStructure(t *testing.T) {
	v := parseString(`
	[web]
	host1

	[web:children]
	nginx
	apache

	[nginx]
	host2

	[apache]
	host3
	`)

	v.assertGroupExists(t, "web")
	v.assertGroupExists(t, "apache")
	v.assertGroupExists(t, "nginx")

	assert(t, len(v.Groups) == 5, "Five groups must present: web, apache, nginx, all, ungrouped")

	v.Groups["web"].assertChildGroupExists(t, "nginx")
	v.Groups["web"].assertChildGroupExists(t, "apache")
	v.Groups["nginx"].assertParentGroupExists(t, "web")
	v.Groups["apache"].assertParentGroupExists(t, "web")

}

func TestHostExpansionFullNumericPattern(t *testing.T) {
	v := parseString(`
	host-[001:015:3]-web:23
	`)
	assert(t, len(v.Hosts) == 5, fmt.Sprintf("There must be 5 hosts in the list, found: %d", len(v.Hosts)))
	v.assertHostExists(t, "host-001-web")
	v.assertHostExists(t, "host-004-web")
	v.assertHostExists(t, "host-007-web")
	v.assertHostExists(t, "host-010-web")
	v.assertHostExists(t, "host-013-web")

	assert(t, v.Hosts["host-007-web"].Port == 23, "host-007-web ports doesn't match")
}

func TestHostExpansionFullAlphabeticPattern(t *testing.T) {
	v := parseString(`
	host-[a:o:3]-web
	`)
	v.assertHostExists(t, "host-a-web")
	v.assertHostExists(t, "host-d-web")
	v.assertHostExists(t, "host-g-web")
	v.assertHostExists(t, "host-j-web")
	v.assertHostExists(t, "host-m-web")

}

func TestHostExpansionShortNumericPattern(t *testing.T) {
	v := parseString(`
	host-[:05]-web
	`)
	assert(t, len(v.Hosts) == 6, fmt.Sprintf("There must be 6 hosts in the list, found: %d", len(v.Hosts)))
	v.assertHostExists(t, "host-00-web")
	v.assertHostExists(t, "host-01-web")
	v.assertHostExists(t, "host-02-web")
	v.assertHostExists(t, "host-03-web")
	v.assertHostExists(t, "host-04-web")
	v.assertHostExists(t, "host-05-web")
}

func TestHostExpansionShortAlphabeticPattern(t *testing.T) {
	v := parseString(`
	host-[a:c]-web
	`)
	assert(t, len(v.Hosts) == 3, fmt.Sprintf("There must be 3 hosts in the list, found: %d", len(v.Hosts)))
	v.assertHostExists(t, "host-a-web")
	v.assertHostExists(t, "host-b-web")
	v.assertHostExists(t, "host-c-web")
}

func TestVariablesPriority(t *testing.T) {
	v := parseString(`
	host-ungrouped-with-x x=a
	host-ungrouped

	[web]
	host-web x=b

	[web:vars]
	x=c

	[web:children]
	nginx

	[nginx:vars]
	x=d

	[nginx]
	host-nginx
	host-nginx-with-x x=e

	[all:vars]
	x=f
	`)

	v.Hosts["host-nginx-with-x"].assertVar(t, "x", "e")
	v.Hosts["host-nginx"].assertVar(t, "x", "d")
	v.Hosts["host-web"].assertVar(t, "x", "b")
	v.Hosts["host-ungrouped-with-x"].assertVar(t, "x", "a")
	v.Hosts["host-ungrouped"].assertVar(t, "x", "f")

}

func TestHostMatching(t *testing.T) {
	inventory := parseString(`
	catfish
	[web:children]
	tomcat         # This is a group!

	[tomcat]
	tomcat
	tomcat-1
	cat
	`)
	hosts := inventory.Match("*cat*")
	assert(t, len(hosts) == 4, fmt.Sprintf("Should be 4, got: %d\n%v", len(hosts), getNames(hosts)))

}

func getNames(hosts []*Host) []string {
	var result []string
	for _, host := range hosts {
		result = append(result, host.Name)
	}
	return result
}
