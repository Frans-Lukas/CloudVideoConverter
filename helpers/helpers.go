package helpers

import "strings"

func ChangeFileExtension(fileName string, extension string) string {
	temp := strings.Split(fileName, ".")[0]
	return temp + "." + extension
}

