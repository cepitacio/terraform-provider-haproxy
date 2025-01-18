package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func GetResourceswithIndexToBeUpdated(d *schema.ResourceData, resourceKey string) ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}, error) {
	var updatedItems, createdItems, deletedItems []map[string]interface{}
	var err error
	oldList, newList := d.GetChange(resourceKey)

	oldItems := oldList.(*schema.Set).List()
	newItems := newList.(*schema.Set).List()

	// Convert oldItems to a map for easier lookup
	oldItemsMap := make(map[int]map[string]interface{})
	sort.Slice(oldItems, func(i, j int) bool {
		mapI := oldItems[i].(map[string]interface{})
		mapJ := oldItems[j].(map[string]interface{})
		return mapI["index"].(int) < mapJ["index"].(int)
	})

	for _, oldItem := range oldItems {
		oldItemMap := oldItem.(map[string]interface{})
		oldIndex := oldItemMap["index"].(int)
		oldItemsMap[oldIndex] = oldItemMap
	}

	// Convert newItems to a map for easier lookup
	newItemsMap := make(map[int]map[string]interface{})
	sort.Slice(newItems, func(i, j int) bool {
		mapI := newItems[i].(map[string]interface{})
		mapJ := newItems[j].(map[string]interface{})
		return mapI["index"].(int) < mapJ["index"].(int)
	})

	for _, newItem := range newItems {
		newItemMap := newItem.(map[string]interface{})
		newIndex := newItemMap["index"].(int)
		newItemsMap[newIndex] = newItemMap
	}

	// Iterate over the new items to determine updates and creations
	for newIndex, newItemMap := range newItemsMap {
		oldItemMap, exists := oldItemsMap[newIndex]

		if exists {
			// Determine if the block has changed
			if !reflect.DeepEqual(oldItemMap, newItemMap) {
				log.Printf("Updating block %d: %v", newIndex, newItemMap)
				updatedItems = append(updatedItems, newItemMap)
			}
		} else {
			// This is a new item that needs to be created
			log.Printf("Creating block %d: %v", newIndex, newItemMap)
			createdItems = append(createdItems, newItemMap)
		}
	}

	// Iterate over the old items to determine deletions
	for oldIndex, oldItemMap := range oldItemsMap {
		_, exists := newItemsMap[oldIndex]
		if !exists {
			// This old item has been removed
			log.Printf("Deleting block %d: %v", oldIndex, oldItemMap)
			deletedItems = append(deletedItems, oldItemMap)
		}
	}
	// }

	if len(updatedItems) > 0 {
		fmt.Println("updatedItems is not empty. Sorting it for later use")
		updatedItems, err = SortItemsByIndex(updatedItems)
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		fmt.Println("updatedItems is empty. Nothing to sort. continuing...")
	}

	if len(createdItems) > 0 {
		fmt.Println("createdItems is not empty. Sorting it for later use")
		createdItems, err = SortItemsByIndex(createdItems)
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		fmt.Println("createdItems is empty. Nothing to sort. continuing...")
	}

	if len(deletedItems) > 0 {
		fmt.Println("deletedItems is not empty. Sorting it for later use")
		deletedItems, err = ReverseSortByIndex(deletedItems)
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		fmt.Println("deletedItems is empty. Nothing to sort. continuing...")
	}

	fmt.Println("Items to be updated ", updatedItems)
	fmt.Println("Items to be created", createdItems)
	fmt.Println("Items to be deleted", deletedItems)
	return updatedItems, createdItems, deletedItems, nil
}

func ProcessUpdateResourceswithIndex(Config interface{}, methodName string, items []map[string]interface{}, transactionID string, parentName string, parentType string) (*http.Response, error) {
	methodValue := reflect.ValueOf(Config).MethodByName(methodName)
	if !methodValue.IsValid() {
		return nil, errors.New("method not found")
	}
	// var resp *http.Response
	for _, item := range items {
		var result []reflect.Value

		// Extract arguments based on the method name
		var args []reflect.Value
		switch methodName {
		case "UpdateAnAclConfiguration":
			index := item["index"].(int)
			payloadJSON, err := json.Marshal(item)
			if err != nil {
				return nil, err
			}
			args = []reflect.Value{
				reflect.ValueOf(index),
				reflect.ValueOf(payloadJSON),
				reflect.ValueOf(transactionID),
				reflect.ValueOf(parentName),
				reflect.ValueOf(parentType),
			}
		case "DeleteAnAclConfiguration":
			index := item["index"].(int)
			args = []reflect.Value{
				reflect.ValueOf(index),
				reflect.ValueOf(transactionID),
				reflect.ValueOf(parentName),
				reflect.ValueOf(parentType),
			}
		case "UpdateAHttpRequestRuleConfiguration":
			index := item["index"].(int)
			payloadJSON, err := json.Marshal(item)
			if err != nil {
				return nil, err
			}
			args = []reflect.Value{
				reflect.ValueOf(index),
				reflect.ValueOf(payloadJSON),
				reflect.ValueOf(transactionID),
				reflect.ValueOf(parentName),
				reflect.ValueOf(parentType),
			}
		case "DeleteAHttpRequestRuleConfiguration":
			index := item["index"].(int)
			args = []reflect.Value{
				reflect.ValueOf(index),
				reflect.ValueOf(transactionID),
				reflect.ValueOf(parentName),
				reflect.ValueOf(parentType),
			}

		case "UpdateATcpRequestRuleConfiguration":
			index := item["index"].(int)
			payloadJSON, err := json.Marshal(item)
			if err != nil {
				return nil, err
			}
			args = []reflect.Value{
				reflect.ValueOf(index),
				reflect.ValueOf(payloadJSON),
				reflect.ValueOf(transactionID),
				reflect.ValueOf(parentName),
				reflect.ValueOf(parentType),
			}
		case "DeleteATcpRequestRuleConfiguration":
			index := item["index"].(int)
			args = []reflect.Value{
				reflect.ValueOf(index),
				reflect.ValueOf(transactionID),
				reflect.ValueOf(parentName),
				reflect.ValueOf(parentType),
			}

		case "UpdateAHttpResponseRuleConfiguration":
			index := item["index"].(int)
			payloadJSON, err := json.Marshal(item)
			if err != nil {
				return nil, err
			}
			args = []reflect.Value{
				reflect.ValueOf(index),
				reflect.ValueOf(payloadJSON),
				reflect.ValueOf(transactionID),
				reflect.ValueOf(parentName),
				reflect.ValueOf(parentType),
			}
		case "DeleteAHttpResponseRuleConfiguration":
			index := item["index"].(int)
			args = []reflect.Value{
				reflect.ValueOf(index),
				reflect.ValueOf(transactionID),
				reflect.ValueOf(parentName),
				reflect.ValueOf(parentType),
			}
		case "DeleteAHttpCheckConfiguration":
			index := item["index"].(int)
			args = []reflect.Value{
				reflect.ValueOf(index),
				reflect.ValueOf(transactionID),
				reflect.ValueOf(parentName),
				reflect.ValueOf(parentType),
			}
		case "UpdateAHttpCheckConfiguration":
			index := item["index"].(int)
			payloadJSON, err := json.Marshal(item)
			if err != nil {
				return nil, err
			}
			args = []reflect.Value{
				reflect.ValueOf(index),
				reflect.ValueOf(payloadJSON),
				reflect.ValueOf(transactionID),
				reflect.ValueOf(parentName),
				reflect.ValueOf(parentType),
			}
		default:
			return nil, errors.New("unsupported method name")
		}

		// Call the method dynamically
		result = methodValue.Call(args)

		// Check for errors
		if len(result) == 2 {
			resp, _ := result[0].Interface().(*http.Response)
			if !result[1].IsNil() {
				return resp, result[1].Interface().(error)
			}
			return resp, nil
		}
	}
	return nil, errors.New("unexpected return values from method call")
}

func GetResourceswithoutIndexToBeUpdated(d *schema.ResourceData, resourceKey string) ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}, error) {
	var updatedItems, createdItems, deletedItems []map[string]interface{}

	// Get old and new values for the resource key
	oldListInterface, newListInterface := d.GetChange(resourceKey)

	// Type assertions to ensure we're working with *schema.Set
	oldList, ok := oldListInterface.(*schema.Set)
	if !ok {
		return nil, nil, nil, fmt.Errorf("expected oldListInterface to be *schema.Set, got %T", oldListInterface)
	}
	newList, ok := newListInterface.(*schema.Set)
	if !ok {
		return nil, nil, nil, fmt.Errorf("expected newListInterface to be *schema.Set, got %T", newListInterface)
	}

	// Convert *schema.Set to []interface{}
	oldItems := oldList.List()
	newItems := newList.List()

	// Convert oldItems and newItems to []map[string]interface{}
	convertToMaps := func(data []interface{}) ([]map[string]interface{}, error) {
		list := make([]map[string]interface{}, 0, len(data))
		for _, item := range data {
			if m, ok := item.(map[string]interface{}); ok {
				list = append(list, m)
			} else {
				return nil, fmt.Errorf("invalid item type found: expected map[string]interface{}, got %T", item)
			}
		}
		return list, nil
	}

	oldMaps, err := convertToMaps(oldItems)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to convert oldItems: %s", err)
	}
	newMaps, err := convertToMaps(newItems)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to convert newItems: %s", err)
	}

	// Helper function to compare maps, focusing on individual keys
	isEqual := func(a, b map[string]interface{}) bool {
		if len(a) != len(b) {
			return false
		}
		for key, valueA := range a {
			if valueB, ok := b[key]; !ok || valueA != valueB {
				return false
			}
		}
		return true
	}

	// Iterate over newItems and compare each bind individually
	for _, newItem := range newMaps {
		found := false
		for _, oldItem := range oldMaps {
			if newItem["name"] == oldItem["name"] {
				if !isEqual(newItem, oldItem) {
					updatedItems = append(updatedItems, newItem)
				}
				found = true
				break
			}
		}

		if !found {
			createdItems = append(createdItems, newItem)
		}
	}

	// Determine deleted items (those in oldItems but not in newItems)
	for _, oldItem := range oldMaps {
		found := false
		for _, newItem := range newMaps {
			if newItem["name"] == oldItem["name"] {
				found = true
				break
			}
		}
		if !found {
			deletedItems = append(deletedItems, oldItem)
		}
	}

	// Log the results for debugging
	logJSON(updatedItems, "Updated Items")
	logJSON(createdItems, "Created Items")
	logJSON(deletedItems, "Deleted Items")

	return updatedItems, createdItems, deletedItems, nil
}

// Helper function to log JSON output
func logJSON(data interface{}, title string) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling %s: %s\n", title, err)
		return
	}
	fmt.Printf("%s:\n%s\n", title, string(jsonData))
}

func ProcessUpdateResourceswithoutIndex(Config interface{}, methodName string, item map[string]interface{}, transactionID string, parentName string, parentType string) (*http.Response, error) {
	fmt.Println("Processing item without index:", item)
	fmt.Println(methodName)
	methodValue := reflect.ValueOf(Config).MethodByName(methodName)
	if !methodValue.IsValid() {
		return nil, errors.New("method not found")
	}

	// for _, item := range items {
	var result []reflect.Value

	// Extract arguments based on the method name
	var args []reflect.Value
	switch methodName {
	case "AddAnAclConfiguration":
		payloadJSON, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		args = []reflect.Value{
			reflect.ValueOf(payloadJSON),
			reflect.ValueOf(transactionID),
			reflect.ValueOf(parentName),
			reflect.ValueOf(parentType),
		}
	case "AddAHttpRequestRuleConfiguration":
		payloadJSON, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		args = []reflect.Value{
			reflect.ValueOf(payloadJSON),
			reflect.ValueOf(transactionID),
			reflect.ValueOf(parentName),
			reflect.ValueOf(parentType),
		}
	case "AddATcpRequestRuleConfiguration":
		payloadJSON, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		args = []reflect.Value{
			reflect.ValueOf(payloadJSON),
			reflect.ValueOf(transactionID),
			reflect.ValueOf(parentName),
			reflect.ValueOf(parentType),
		}
	case "AddAHttpResponseRuleConfiguration":
		payloadJSON, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		args = []reflect.Value{
			reflect.ValueOf(payloadJSON),
			reflect.ValueOf(transactionID),
			reflect.ValueOf(parentName),
			reflect.ValueOf(parentType),
		}
	case "AddABindConfiguration":
		payloadJSON, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		args = []reflect.Value{
			reflect.ValueOf(payloadJSON),
			reflect.ValueOf(transactionID),
			reflect.ValueOf(parentName),
			reflect.ValueOf(parentType),
		}
	case "AddAHttpCheckConfiguration":
		payloadJSON, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		args = []reflect.Value{
			reflect.ValueOf(payloadJSON),
			reflect.ValueOf(transactionID),
			reflect.ValueOf(parentName),
			reflect.ValueOf(parentType),
		}
	default:
		return nil, errors.New("unsupported method name")
	}

	// Call the method dynamically
	result = methodValue.Call(args)

	if len(result) == 2 {
		resp, _ := result[0].Interface().(*http.Response) // Safely cast to *http.Response
		if !result[1].IsNil() {                           // Check if error is not nil
			return resp, result[1].Interface().(error)
		}
		return resp, nil
	}

	return nil, errors.New("unexpected return values from method call")
}

func ProcessUpdateResourceswithoutIndexAndName(Config interface{}, methodName string, item map[string]interface{}, transactionID string, frontendName string, parentName string, parentType string) (*http.Response, error) {
	fmt.Println("Processing item without index:", item)
	fmt.Println(methodName)
	methodValue := reflect.ValueOf(Config).MethodByName(methodName)
	if !methodValue.IsValid() {
		return nil, errors.New("method not found")
	}

	// for _, item := range items {
	var result []reflect.Value

	// Extract arguments based on the method name
	var args []reflect.Value
	switch methodName {

	case "UpdateABindConfiguration":
		payloadJSON, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		args = []reflect.Value{
			reflect.ValueOf(frontendName),
			reflect.ValueOf(payloadJSON),
			reflect.ValueOf(transactionID),
			reflect.ValueOf(parentName),
			reflect.ValueOf(parentType),
		}
	case "DeleteABindConfiguration":
		args = []reflect.Value{
			reflect.ValueOf(frontendName),
			reflect.ValueOf(transactionID),
			reflect.ValueOf(parentName),
			reflect.ValueOf(parentType),
		}
	default:
		return nil, errors.New("unsupported method name")
	}

	// Call the method dynamically
	result = methodValue.Call(args)

	if len(result) == 2 {
		resp, _ := result[0].Interface().(*http.Response) // Safely cast to *http.Response
		if !result[1].IsNil() {                           // Check if error is not nil
			return resp, result[1].Interface().(error)
		}
		return resp, nil
	}

	return nil, errors.New("unexpected return values from method call")
}
