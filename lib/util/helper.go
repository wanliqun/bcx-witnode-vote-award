package util

func SearchStringSlice(s []string, e string) int {
	for i, v := range s {
		if v == e {
			return i
		}
	}
	return -1
}

func RemoveStringSliceAt(s []string, index int) []string {
	if index < 0 || index >= len(s) {
		return s
	}

	if index == len(s) -1 {
		return s[:index]
	}

	return append(s[:index], s[index+1:]...)
}