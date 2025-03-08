package util

import "strings"

// normalizeServiceName removes the .service suffix if present
func NormalizeServiceName(serviceName string) string {
	return strings.TrimSpace(strings.TrimSuffix(serviceName, ".service"))
}
