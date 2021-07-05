package aini

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddVars(t *testing.T) {
	v, err := ParseFile("test_data/inventory")
	assert.Nil(t, err)

	assert.Equal(t, "present", v.Groups["web"].Vars["web_inventory_string_var"])
	assert.Equal(t, "should be overwritten", v.Groups["web"].Vars["web_string_var"])

	assert.Equal(t, "present", v.Hosts["host1"].Vars["host1_inventory_string_var"])
	assert.Equal(t, "should be overwritten", v.Hosts["host1"].Vars["host1_string_var"])

	err = v.AddVars("test_data")
	assert.Nil(t, err)

	assert.Equal(t, "1", v.Groups["web"].Vars["web_int_var"])
	assert.Equal(t, "string", v.Groups["web"].Vars["web_string_var"])
	assert.Equal(t, "{\"this\":{\"is\":\"object\"}}", v.Groups["web"].Vars["web_object_var"])
	assert.Equal(t, "present", v.Groups["web"].Vars["web_inventory_string_var"])

	assert.Equal(t, "1", v.Groups["nginx"].Vars["nginx_int_var"])
	assert.Equal(t, "string", v.Groups["nginx"].Vars["nginx_string_var"])
	assert.Equal(t, "true", v.Groups["nginx"].Vars["nginx_bool_var"])
	assert.Equal(t, "{\"this\":{\"is\":\"object\"}}", v.Groups["nginx"].Vars["nginx_object_var"])

	assert.Equal(t, "1", v.Hosts["host1"].Vars["host1_int_var"])
	assert.Equal(t, "string", v.Hosts["host1"].Vars["host1_string_var"])
	assert.Equal(t, "{\"this\":{\"is\":\"object\"}}", v.Hosts["host1"].Vars["host1_object_var"])
	assert.Equal(t, "present", v.Hosts["host1"].Vars["host1_inventory_string_var"])

	assert.Equal(t, "1", v.Hosts["host2"].Vars["host2_int_var"])
	assert.Equal(t, "string", v.Hosts["host2"].Vars["host2_string_var"])
	assert.Equal(t, "{\"this\":{\"is\":\"object\"}}", v.Hosts["host2"].Vars["host2_object_var"])

	assert.NotContains(t, v.Groups, "tomcat")
	assert.NotContains(t, v.Hosts, "host7")
}

func TestAddVarsLowerCased(t *testing.T) {
	v, err := ParseFile("test_data/inventory")
	assert.Nil(t, err)

	v.HostsToLower()
	v.GroupsToLower()
	v.AddVarsLowerCased("test_data")

	assert.Contains(t, v.Groups, "tomcat")
	assert.Contains(t, v.Hosts, "host7")
	assert.Equal(t, "string", v.Groups["tomcat"].Vars["tomcat_string_var"])
	assert.Equal(t, "string", v.Hosts["host7"].Vars["host7_string_var"])
}
