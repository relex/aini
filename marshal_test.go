package aini

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const minMarshalInventory = `[Animals]
ET

[Animals:children]
Cats

[Cats]
Lion
`

//go:embed marshal_test_inventory.json
var minMarshalJSON string

func TestMarshalJSON(t *testing.T) {
	v, err := ParseString(minMarshalInventory)
	assert.Nil(t, err)

	j, err := json.MarshalIndent(v, "", "    ")
	assert.Nil(t, err)
	assert.Equal(t, minMarshalJSON, string(j))

	t.Run("unmarshal", func(t *testing.T) {
		var v2 InventoryData
		assert.Nil(t, json.Unmarshal(j, &v2))
		assert.Equal(t, v.Hosts["Lion"], v2.Hosts["Lion"])
		assert.Equal(t, v.Groups["Cats"], v2.Groups["Cats"])
	})
}

func TestMarshalWithVars(t *testing.T) {
	v, err := ParseFile("test_data/inventory")
	assert.Nil(t, err)

	v.HostsToLower()
	v.GroupsToLower()
	v.AddVarsLowerCased("test_data")

	j, err := json.MarshalIndent(v, "", "    ")
	assert.Nil(t, err)

	t.Run("unmarshal", func(t *testing.T) {
		var v2 InventoryData
		assert.Nil(t, json.Unmarshal(j, &v2))
		assert.Equal(t, v.Hosts["host1"], v2.Hosts["host1"])
		assert.Equal(t, v.Groups["tomcat"], v2.Groups["tomcat"])
	})
}
