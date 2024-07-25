package aini

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiPatterns(t *testing.T) {
	var ok bool
	var err error

	ok, err = MatchNamesByPatterns([]string{"myhost", "web", "tmp"}, []string{"all"})
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = MatchNamesByPatterns([]string{"myhost", "web", "tmp"}, []string{"all", "tmp"})
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = MatchNamesByPatterns([]string{"myhost", "web", "tmp"}, []string{"all", "!tmp"})
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = MatchNamesByPatterns([]string{"myhost", "web", "tmp"}, []string{"all", "&tmp"})
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = MatchNamesByPatterns([]string{"myhost", "web", "tmp"}, []string{"web", "myhost"})
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = MatchNamesByPatterns([]string{"myhost", "web", "tmp"}, []string{"web", "myhost", "&hello"})
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestGroupsMatching(t *testing.T) {
	v := parseString(t, `
	host1
	host2
	[myGroup1]
	host1
	[myGroup2]
	host1
	[groupForCats]
	host1
	`)

	groups, err := v.MatchGroups("*Group*")
	assert.Nil(t, err)
	assert.Contains(t, groups, "myGroup1")
	assert.Contains(t, groups, "myGroup2")
	assert.Len(t, groups, 2)

	groups, err = v.Hosts["host1"].MatchGroups("*Group*")
	assert.Nil(t, err)
	assert.Contains(t, groups, "myGroup1")
	assert.Contains(t, groups, "myGroup2")
	assert.Len(t, groups, 2)

	ok, err := v.Hosts["host1"].MatchPatterns([]string{"host[12]", "myGroup*", "&groupForCats"})
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = v.Hosts["host1"].MatchPatterns([]string{"host[12]", "myGroup*", "&groupForCats", "&notHere"})
	assert.Nil(t, err)
	assert.False(t, ok)

	ok, err = v.Hosts["host1"].MatchPatterns([]string{"host[12]", "myGroup*", "!groupFor*"})
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestHostsMatching(t *testing.T) {
	v := parseString(t, `
	myHost1
	otherHost2
	[group1]
	myHost1
	[group2]
	myHost1
	myHost2
	[group3:children]
	group2
	`)

	hosts, err := v.MatchHosts("my*")
	assert.Nil(t, err)
	assert.Contains(t, hosts, "myHost1")
	assert.Contains(t, hosts, "myHost2")
	assert.Len(t, hosts, 2)

	hosts, err = v.Groups["group1"].MatchHosts("*my*")
	assert.Nil(t, err)
	assert.Contains(t, hosts, "myHost1")
	assert.Len(t, hosts, 1)

	hosts, err = v.Groups["group2"].MatchHosts("*my*")
	assert.Nil(t, err)
	assert.Contains(t, hosts, "myHost1")
	assert.Contains(t, hosts, "myHost2")
	assert.Len(t, hosts, 2)

	hosts, err = v.MatchHostsByPatterns("group3:otherHost[0-9]:!group1")
	assert.Nil(t, err)
	assert.Len(t, hosts, 2)
	assert.NotContains(t, hosts, "myHost1")
	assert.Contains(t, hosts, "myHost2")
	assert.Contains(t, hosts, "otherHost2")
}

func TestVarsMatching(t *testing.T) {
	v := parseString(t, `
	host1 myHostVar=myHostVarValue otherHostVar=otherHostVarValue
	
	[group1]
	host1

	[group1:vars]
	myGroupVar=myGroupVarValue
	otherGroupVar=otherGroupVarValue
	`)
	group := v.Groups["group1"]
	vars, err := group.MatchVars("my*")
	assert.Nil(t, err)
	assert.Contains(t, vars, "myGroupVar")
	assert.Len(t, vars, 1)
	assert.Equal(t, "myGroupVarValue", vars["myGroupVar"])

	host := v.Hosts["host1"]
	vars, err = host.MatchVars("my*")
	assert.Nil(t, err)
	assert.Contains(t, vars, "myHostVar")
	assert.Contains(t, vars, "myGroupVar")
	assert.Len(t, vars, 2)
	assert.Equal(t, "myHostVarValue", vars["myHostVar"])
	assert.Equal(t, "myGroupVarValue", vars["myGroupVar"])
}
