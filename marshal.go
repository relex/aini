package aini

import (
	"encoding/json"

	"github.com/samber/lo"
	"golang.org/x/exp/maps"
)

type alwaysNil interface{} // to hold place for Group and Host references; must be nil in serialized form

func (group *Group) MarshalJSON() ([]byte, error) {
	type groupWithoutCustomMarshal Group

	return json.Marshal(&struct {
		groupWithoutCustomMarshal
		Hosts         map[string]alwaysNil
		Children      map[string]alwaysNil
		Parents       map[string]alwaysNil
		DirectParents map[string]alwaysNil
	}{
		groupWithoutCustomMarshal: groupWithoutCustomMarshal(*group),
		Hosts:                     makeNilValueMap(group.Hosts),
		Children:                  makeNilValueMap(group.Children),
		Parents:                   makeNilValueMap(group.Parents),
		DirectParents:             makeNilValueMap(group.DirectParents),
	})
}

func (host *Host) MarshalJSON() ([]byte, error) {
	type hostWithoutCustomMarshal Host

	return json.Marshal(&struct {
		hostWithoutCustomMarshal
		Groups       map[string]alwaysNil
		DirectGroups map[string]alwaysNil
	}{
		hostWithoutCustomMarshal: hostWithoutCustomMarshal(*host),
		Groups:                   makeNilValueMap(host.Groups),
		DirectGroups:             makeNilValueMap(host.DirectGroups),
	})
}

func makeNilValueMap[K comparable, V any](m map[K]*V) map[K]alwaysNil {
	return lo.MapValues(m, func(_ *V, _ K) alwaysNil { return nil })
}

func (inventory *InventoryData) UnmarshalJSON(data []byte) error {
	type inventoryWithoutCustomUnmarshal InventoryData
	var rawInventory inventoryWithoutCustomUnmarshal
	if err := json.Unmarshal(data, &rawInventory); err != nil {
		return err
	}
	// rawInventory's Groups and Hosts should now contain all properties,
	// except child group maps and host maps are filled with original keys and null values

	// reassign child groups and hosts to reference rawInventory.Hosts and .Groups

	for _, group := range rawInventory.Groups {
		group.Hosts = lo.PickByKeys(rawInventory.Hosts, maps.Keys(group.Hosts))
		group.Children = lo.PickByKeys(rawInventory.Groups, maps.Keys(group.Children))
		group.Parents = lo.PickByKeys(rawInventory.Groups, maps.Keys(group.Parents))
		group.DirectParents = lo.PickByKeys(rawInventory.Groups, maps.Keys(group.DirectParents))
	}

	for _, host := range rawInventory.Hosts {
		host.Groups = lo.PickByKeys(rawInventory.Groups, maps.Keys(host.Groups))
		host.DirectGroups = lo.PickByKeys(rawInventory.Groups, maps.Keys(host.DirectGroups))
	}

	inventory.Groups = rawInventory.Groups
	inventory.Hosts = rawInventory.Hosts
	return nil
}
