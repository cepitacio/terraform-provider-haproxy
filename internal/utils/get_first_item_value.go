package utils

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func GetFirstItemValue(getFunc func(key string) interface{}, key string) interface{} {
	item := getFunc(key)
	if item == nil {
		return nil
	}

	list, ok := item.(*schema.Set)
	if !ok || list.Len() == 0 {
		return nil
	}
	println(list)
	return list.List()[0]
}
