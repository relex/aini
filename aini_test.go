package aini

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func parseString(t *testing.T, input string) *InventoryData {
	v, err := ParseString(input)
	assert.Nil(t, err, fmt.Sprintf("Error occurred while parsing: %s", err))
	return v
}

func TestBelongToBasicGroups(t *testing.T) {
	v := parseString(t, `
	host1:2221 # Comments
	[web]      # should
	host2      # be
	           # ignored
	`)

	assert.Len(t, v.Hosts, 2, "Exactly two hosts expected")
	assert.Len(t, v.Groups, 3, "Expected three groups: web, all and ungrouped")

	assert.Contains(t, v.Groups, "web")
	assert.Contains(t, v.Groups, "all")
	assert.Contains(t, v.Groups, "ungrouped")

	assert.Contains(t, v.Hosts, "host1")
	assert.Len(t, v.Hosts["host1"].Groups, 2, "Host1 must belong to two groups: ungrouped and all")
	assert.NotNil(t, 2, v.Hosts["host1"].Groups["all"], "Host1 must belong to two groups: ungrouped and all")
	assert.NotNil(t, 2, v.Hosts["host1"].Groups["ungrouped"], "Host1 must belong to ungrouped group")

	assert.Contains(t, v.Hosts, "host2")
	assert.Len(t, v.Hosts["host2"].Groups, 2, "Host2 must belong to two groups: ungrouped and all")
	assert.NotNil(t, 2, v.Hosts["host2"].Groups["all"], "Host2 must belong to two groups: ungrouped and all")
	assert.NotNil(t, 2, v.Hosts["host2"].Groups["ungrouped"], "Host2 must belong to ungrouped group")

	assert.Equal(t, 2, len(v.Groups["all"].Hosts), "Group all must contain two hosts")
	assert.Contains(t, v.Groups["all"].Hosts, "host1")
	assert.Contains(t, v.Groups["all"].Hosts, "host2")

	assert.Len(t, v.Groups["web"].Hosts, 1, "Group web must contain one host")
	assert.Contains(t, v.Groups["web"].Hosts, "host2")

	assert.Len(t, v.Groups["ungrouped"].Hosts, 1, "Group ungrouped must contain one host")
	assert.Contains(t, v.Groups["ungrouped"].Hosts, "host1")
	assert.NotContains(t, v.Groups["ungrouped"].Hosts, "host2")

	assert.Equal(t, 2221, v.Hosts["host1"].Port, "Host1 port is set")
	assert.Equal(t, 22, v.Hosts["host2"].Port, "Host2 port is set")
}

func TestGroupStructure(t *testing.T) {
	v := parseString(t, `
	host5

	[web:children]
	nginx
	apache

	[web]
	host1
	host2

	[nginx]
	host1
	host3
	host4

	[apache]
	host5
	host6
	`)

	assert.Contains(t, v.Groups, "web")
	assert.Contains(t, v.Groups, "apache")
	assert.Contains(t, v.Groups, "nginx")

	assert.Len(t, v.Groups, 5, "Five groups must present: web, apache, nginx, all, ungrouped")

	assert.Contains(t, v.Groups["web"].Children, "nginx")
	assert.Contains(t, v.Groups["web"].Children, "apache")
	assert.Contains(t, v.Groups["nginx"].Parents, "web")
	assert.Contains(t, v.Groups["apache"].Parents, "web")

	assert.Contains(t, v.Groups["web"].Hosts, "host1")
	assert.Contains(t, v.Groups["web"].Hosts, "host2")
	assert.Contains(t, v.Groups["web"].Hosts, "host3")
	assert.Contains(t, v.Groups["web"].Hosts, "host4")
	assert.Contains(t, v.Groups["web"].Hosts, "host5")

	assert.Contains(t, v.Groups["nginx"].Hosts, "host1")

	assert.Contains(t, v.Hosts["host1"].Groups, "web")
	assert.Contains(t, v.Hosts["host1"].Groups, "nginx")

	assert.Empty(t, v.Groups["ungrouped"].Hosts)
}

func TestGroupNotExplicitlyDefined(t *testing.T) {
	v := parseString(t, `
	[web:children]
	nginx

	[nginx]
	host1
	`)

	assert.Contains(t, v.Groups, "web")
	assert.Contains(t, v.Groups, "nginx")

	assert.Len(t, v.Groups, 4, "Four groups must present: web, nginx, all, ungrouped")

	assert.Contains(t, v.Groups["web"].Children, "nginx")
	assert.Contains(t, v.Groups["nginx"].Parents, "web")

	assert.Contains(t, v.Groups["web"].Hosts, "host1")
	assert.Contains(t, v.Groups["nginx"].Hosts, "host1")

	assert.Contains(t, v.Hosts["host1"].Groups, "web")
	assert.Contains(t, v.Hosts["host1"].Groups, "nginx")

	assert.Empty(t, v.Groups["ungrouped"].Hosts, "Group ungrouped should be empty")
}

func TestHostExpansionFullNumericPattern(t *testing.T) {
	v := parseString(t, `
	host-[001:015:3]-web:23
	`)

	assert.Len(t, v.Hosts, 5)
	assert.Contains(t, v.Hosts, "host-001-web")
	assert.Contains(t, v.Hosts, "host-004-web")
	assert.Contains(t, v.Hosts, "host-007-web")
	assert.Contains(t, v.Hosts, "host-010-web")
	assert.Contains(t, v.Hosts, "host-013-web")

	for _, host := range v.Hosts {
		assert.Equalf(t, 23, host.Port, "%s port is set", host.Name)
	}
}

func TestHostExpansionFullAlphabeticPattern(t *testing.T) {
	v := parseString(t, `
	host-[a:o:3]-web
	`)

	assert.Len(t, v.Hosts, 5)
	assert.Contains(t, v.Hosts, "host-a-web")
	assert.Contains(t, v.Hosts, "host-d-web")
	assert.Contains(t, v.Hosts, "host-g-web")
	assert.Contains(t, v.Hosts, "host-j-web")
	assert.Contains(t, v.Hosts, "host-m-web")
}

func TestHostExpansionShortNumericPattern(t *testing.T) {
	v := parseString(t, `
	host-[:05]-web
	`)
	assert.Len(t, v.Hosts, 6)
	assert.Contains(t, v.Hosts, "host-00-web")
	assert.Contains(t, v.Hosts, "host-01-web")
	assert.Contains(t, v.Hosts, "host-02-web")
	assert.Contains(t, v.Hosts, "host-03-web")
	assert.Contains(t, v.Hosts, "host-04-web")
	assert.Contains(t, v.Hosts, "host-05-web")
}

func TestHostExpansionShortAlphabeticPattern(t *testing.T) {
	v := parseString(t, `
	host-[a:c]-web
	`)
	assert.Len(t, v.Hosts, 3)
	assert.Contains(t, v.Hosts, "host-a-web")
	assert.Contains(t, v.Hosts, "host-b-web")
	assert.Contains(t, v.Hosts, "host-c-web")
}

func TestHostExpansionMultiplePatterns(t *testing.T) {
	v := parseString(t, `
	host-[1:2]-[a:b]-web
	`)
	assert.Len(t, v.Hosts, 4)
	assert.Contains(t, v.Hosts, "host-1-a-web")
	assert.Contains(t, v.Hosts, "host-1-b-web")
	assert.Contains(t, v.Hosts, "host-2-a-web")
	assert.Contains(t, v.Hosts, "host-2-b-web")
}

func TestVariablesPriority(t *testing.T) {
	v := parseString(t, `
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

	assert.Equal(t, "a", v.Hosts["host-ungrouped-with-x"].Vars["x"])
	assert.Equal(t, "b", v.Hosts["host-web"].Vars["x"])
	assert.Equal(t, "c", v.Groups["web"].Vars["x"])
	assert.Equal(t, "d", v.Hosts["host-nginx"].Vars["x"])
	assert.Equal(t, "e", v.Hosts["host-nginx-with-x"].Vars["x"])
	assert.Equal(t, "f", v.Hosts["host-ungrouped"].Vars["x"])
}

func TestHostsToLower(t *testing.T) {
	v := parseString(t, `
	CatFish
	[web:children]
	TomCat

	[TomCat]
	TomCat
	tomcat-1
	cat
	`)
	assert.Contains(t, v.Hosts, "CatFish")
	assert.Contains(t, v.Groups["ungrouped"].Hosts, "CatFish")
	assert.Contains(t, v.Hosts, "TomCat")

	v.HostsToLower()

	assert.NotContains(t, v.Hosts, "CatFish")
	assert.Contains(t, v.Hosts, "catfish")
	assert.Equal(t, "catfish", v.Hosts["catfish"].Name, "Host catfish should have a matching name")

	assert.NotContains(t, v.Hosts, "TomCat")
	assert.Contains(t, v.Hosts, "tomcat")
	assert.Equal(t, "tomcat", v.Hosts["tomcat"].Name, "Host tomcat should have a matching name")

	assert.NotContains(t, v.Groups["ungrouped"].Hosts, "CatFish")
	assert.Contains(t, v.Groups["ungrouped"].Hosts, "catfish")
	assert.NotContains(t, v.Groups["web"].Hosts, "TomCat")
	assert.Contains(t, v.Groups["web"].Hosts, "tomcat")
}

func TestGroupsToLower(t *testing.T) {
	v := parseString(t, `
	[Web]
	CatFish

	[Web:children]
	TomCat

	[TomCat]
	TomCat
	tomcat-1
	cat
	`)
	assert.Contains(t, v.Groups, "Web")
	assert.Contains(t, v.Groups, "TomCat")
	v.GroupsToLower()
	assert.NotContains(t, v.Groups, "Web")
	assert.NotContains(t, v.Groups, "TomCat")
	assert.Contains(t, v.Groups, "web")
	assert.Contains(t, v.Groups, "tomcat")

	assert.Equal(t, "web", v.Groups["web"].Name, "Group web should have matching name")
	assert.Contains(t, v.Groups["web"].Children, "tomcat")
	assert.Contains(t, v.Groups["web"].Hosts, "TomCat")

	assert.Equal(t, "tomcat", v.Groups["tomcat"].Name, "Group tomcat should have matching name")
	assert.Contains(t, v.Groups["tomcat"].Hosts, "TomCat")
	assert.Contains(t, v.Groups["tomcat"].Hosts, "tomcat-1")
	assert.Contains(t, v.Groups["tomcat"].Hosts, "cat")
}

func TestGroupsAndHostsToLower(t *testing.T) {
	v := parseString(t, `
	[Web]
	CatFish

	[Web:children]
	TomCat

	[TomCat]
	TomCat
	tomcat-1
	`)
	assert.Contains(t, v.Groups, "Web")
	assert.Contains(t, v.Groups, "TomCat")

	assert.Contains(t, v.Hosts, "CatFish")
	assert.Contains(t, v.Hosts, "TomCat")
	assert.Contains(t, v.Hosts, "tomcat-1")

	v.GroupsToLower()
	v.HostsToLower()

	assert.NotContains(t, v.Groups, "Web")
	assert.NotContains(t, v.Groups, "TomCat")
	assert.Contains(t, v.Groups, "web")
	assert.Contains(t, v.Groups, "tomcat")

	assert.NotContains(t, v.Hosts, "CatFish")
	assert.NotContains(t, v.Hosts, "TomCat")
	assert.Contains(t, v.Hosts, "catfish")
	assert.Contains(t, v.Hosts, "tomcat")
	assert.Contains(t, v.Hosts, "tomcat-1")

	assert.Contains(t, v.Groups["web"].Hosts, "catfish")
	assert.Contains(t, v.Groups["web"].Children, "tomcat")
	assert.Contains(t, v.Groups["tomcat"].Hosts, "tomcat")
	assert.Contains(t, v.Groups["tomcat"].Hosts, "tomcat-1")
}

func TestVariablesEscaping(t *testing.T) {
	v := parseString(t, `
	host ansible_ssh_common_args="-o ProxyCommand='ssh -W %h:%p somehost'" other_var_same_value="-o ProxyCommand='ssh -W %h:%p somehost'" # comment
	`)
	assert.Contains(t, v.Hosts, "host")
	assert.Equal(t, "-o ProxyCommand='ssh -W %h:%p somehost'", v.Hosts["host"].Vars["ansible_ssh_common_args"])
	assert.Equal(t, "-o ProxyCommand='ssh -W %h:%p somehost'", v.Hosts["host"].Vars["other_var_same_value"])
}

func TestComments(t *testing.T) {
	v := parseString(t, `
	catfish        # I'm a comment
	# Whole-line comment
	[web:children] # Look, there is a cat in comment!
	tomcat         # This is a group!
	 # Whole-line comment with a leading space
	[tomcat]       # And here is another cat 🐈
	tomcat         # Host comment
	tomcat-1 # Small indention comment
	cat                                           # Big indention comment
	`)
	assert.Contains(t, v.Groups, "web")
	assert.Contains(t, v.Groups, "tomcat")
	assert.Contains(t, v.Groups["web"].Children, "tomcat")

	assert.Contains(t, v.Hosts, "tomcat")
	assert.Contains(t, v.Hosts, "tomcat-1")
	assert.Contains(t, v.Hosts, "cat")
	assert.Contains(t, v.Groups["tomcat"].Hosts, "tomcat")
	assert.Contains(t, v.Groups["tomcat"].Hosts, "tomcat-1")
	assert.Contains(t, v.Groups["tomcat"].Hosts, "cat")
	assert.Contains(t, v.Hosts, "catfish")
	assert.Contains(t, v.Groups["ungrouped"].Hosts, "catfish")
}

func TestHostMatching(t *testing.T) {
	v := parseString(t, `
	catfish
	[web:children] # Look, there is a cat in comment!
	tomcat         # This is a group!

	[tomcat]       # And here is another cat 🐈
	tomcat
	tomcat-1
	cat
	`)
	hosts := v.Match("*cat*")
	assert.Len(t, hosts, 4)
}

func TestHostMapListValues(t *testing.T) {
	v := parseString(t, `
	host1
	host2
	host3
	`)

	hosts := HostMapListValues(v.Hosts)
	assert.Len(t, hosts, 3)
	for _, v := range hosts {
		assert.Contains(t, hosts, v)
	}
}

func TestGroupMapListValues(t *testing.T) {
	v := parseString(t, `
	[group1]
	[group2]
	[group3]
	`)

	groups := GroupMapListValues(v.Groups)
	assert.Len(t, groups, 5)
	for _, v := range groups {
		assert.Contains(t, groups, v)
	}
}
