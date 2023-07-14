package utils

import "regexp"

func IsComponentKey(key string) bool {
	regex := regexp.MustCompile(`^targets\.[^.]+\.[^.]+`)
	return regex.MatchString(key)
}
