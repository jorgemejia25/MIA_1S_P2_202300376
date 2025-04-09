package utils

import (
	"strings"

	"disk.simulator.com/m/v2/internal/disk/types/structures/authentication"
)

func FindLastUserInFile(fileText string) (*authentication.User, int) {
	var lastUser *authentication.User
	lastIndex := -1

	lines := strings.Split(fileText, "\n")
	for i, line := range lines {
		data := strings.Split(line, ",")
		if len(data) > 1 && data[1] == "U" {
			lastUser = &authentication.User{
				UID:      data[0],
				Group:    data[2],
				Username: data[3],
				Password: data[4],
			}
			lastIndex = i
		}
	}
	return lastUser, lastIndex
}
