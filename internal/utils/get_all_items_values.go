package utils

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func GetAllItemsValues(getFunc func(key string) interface{}, key string) ([]map[string]interface{}, error) {
	item := getFunc(key)
	if item == nil {
		return nil, fmt.Errorf("items retrieved from getFunc is nil")
	}

	list, ok := item.(*schema.Set)
	if !ok {
		return nil, fmt.Errorf("items is not of type *schema.Set")
	}

	var items []map[string]interface{}
	for _, v := range list.List() {
		if itemMap, ok := v.(map[string]interface{}); ok {
			items = append(items, itemMap)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		indexI, _ := items[i]["index"].(int)
		indexJ, _ := items[j]["index"].(int)
		return indexI < indexJ
	})

	return items, nil
}
