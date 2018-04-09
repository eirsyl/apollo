package utils

// MapKeysToString extracts the keys from a map
func MapKeysToString(m map[string]bool) (res []string) {
	for key := range m {
		res = append(res, key)
	}
	return
}

// MapKeysToInt extracts the keys from a map
func MapKeysToInt(m map[int]bool) (res []int) {
	for key := range m {
		res = append(res, key)
	}
	return
}

// StringListToMap creates a map[string]bool based on a []string
func StringListToMap(list []string) (res map[string]bool) {
	res = make(map[string]bool)
	for _, element := range list {
		res[element] = true
	}
	return
}

// RemoveFromStringList removes a string from a list of strings
func RemoveFromStringList(list []string, element string) []string {
	var res []string
	for _, s := range list {
		if s != element {
			res = append(res, s)
		}
	}
	return res
}
