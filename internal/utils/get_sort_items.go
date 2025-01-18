package utils

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func FetchAndSortSchemaItemsByIndex(getFunc func(key string) interface{}, key string) ([]map[string]interface{}, error) {
	// Retrieve items from getFunc
	item := getFunc(key)
	if item == nil {
		return nil, fmt.Errorf("items retrieved from getFunc is nil")
	}

	// Type assertion to *schema.Set
	list, ok := item.(*schema.Set)
	if !ok {
		return nil, fmt.Errorf("items is not of type *schema.Set")
	}

	// Convert items to a slice of maps
	var items []map[string]interface{}
	for _, v := range list.List() {
		if itemMap, ok := v.(map[string]interface{}); ok {
			items = append(items, itemMap)
		}
	}

	// Sort items based on "Index" key
	sort.Slice(items, func(i, j int) bool {
		indexI, _ := items[i]["index"].(int)
		indexJ, _ := items[j]["index"].(int)
		return indexI < indexJ
	})
	jsonData1, err := json.MarshalIndent(items, "", "  ")
	fmt.Println("Sorted schema items by index", string(jsonData1), err)
	return items, nil
}
func FetchAndSortSchemaItemsByIndexList(getFunc func(key string) interface{}, key string) ([]map[string]interface{}, error) {
	// Retrieve the items using the provided getFunc
	itemsRaw := getFunc(key)
	if itemsRaw == nil {
		return nil, fmt.Errorf("items retrieved from getFunc are nil")
	}

	// Ensure it's a slice of interfaces
	itemsSlice, ok := itemsRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("items is not of type []interface{}, but %T", itemsRaw)
	}

	// Convert items to a slice of maps
	var items []map[string]interface{}
	for _, item := range itemsSlice {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("item is not of type map[string]interface{}, but %T", item)
		}
		items = append(items, itemMap)
	}

	// Sort items based on the "index" key
	sort.Slice(items, func(i, j int) bool {
		indexI, _ := items[i]["index"].(int)
		indexJ, _ := items[j]["index"].(int)
		return indexI < indexJ
	})

	// Debug: Print the sorted items
	jsonData, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling sorted items: %v\n", err)
	} else {
		fmt.Printf("Sorted Items: %s\n", jsonData)
	}

	return items, nil
}

func FetchAndSortSchemaItemsByIndexReverse(getFunc func(key string) interface{}, key string) ([]map[string]interface{}, error) {
	// Retrieve items from getFunc
	item := getFunc(key)
	if item == nil {
		return nil, fmt.Errorf("items retrieved from getFunc is nil")
	}

	// Type assertion to *schema.Set
	list, ok := item.(*schema.Set)
	if !ok {
		return nil, fmt.Errorf("items is not of type *schema.Set")
	}

	// Convert items to a slice of maps
	var items []map[string]interface{}
	for _, v := range list.List() {
		if itemMap, ok := v.(map[string]interface{}); ok {
			items = append(items, itemMap)
		}
	}

	// Reverse sort items based on "Index" key
	sort.Slice(items, func(i, j int) bool {
		indexI, _ := items[i]["index"].(int)
		indexJ, _ := items[j]["index"].(int)
		return indexI > indexJ // Reverse the comparison for descending order
	})

	// Optionally print the sorted result
	jsonData1, err := json.MarshalIndent(items, "", "  ")
	fmt.Println(string(jsonData1), err)

	return items, nil
}

// SortItemsByIndex sorts the items in ascending order by index
func SortItemsByIndex(items []map[string]interface{}) ([]map[string]interface{}, error) {
	// Sort the items by the "index" field
	sort.Slice(items, func(i, j int) bool {
		return items[i]["index"].(int) < items[j]["index"].(int)
	})
	return items, nil
}

// ReverseSortByIndex sorts the items in descending order by index
func ReverseSortByIndex(items []map[string]interface{}) ([]map[string]interface{}, error) {
	// Sort the items by the "index" field in reverse order
	sort.Slice(items, func(i, j int) bool {
		return items[i]["index"].(int) > items[j]["index"].(int)
	})
	return items, nil
}
