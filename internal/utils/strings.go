package utils

// TruncateID safely truncates an ID string to specified length
// Returns the first maxLen characters if the string is longer
func TruncateID(id string, maxLen int) string {
	if len(id) <= maxLen {
		return id
	}
	return id[:maxLen]
}
