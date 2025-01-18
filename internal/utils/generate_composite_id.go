package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

func GenerateCompositeID(items []map[string]interface{}) string {
	// Collect all `acl_name` values in sorted order
	var names []string
	for _, acl := range items {
		if name, ok := acl["acl_name"].(string); ok {
			names = append(names, name)
		}
	}

	// Sort the names for consistency
	sort.Strings(names)

	// Create a composite string of all names
	composite := strings.Join(names, ",")

	// Generate a SHA-256 hash of the composite string
	hash := sha256.Sum256([]byte(composite))
	return hex.EncodeToString(hash[:])
}
