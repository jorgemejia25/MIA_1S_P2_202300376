package utils

import (
	"strings"

	"disk.simulator.com/m/v2/internal/disk/types/structures/authentication"
)

func FindUserInFile(fileText string, user string) (*authentication.User, int) {

	lines := strings.Split(fileText, "\n")

	for i, line := range lines {
		// split line by comma
		data := strings.Split(line, ",")

		if len(data) > 1 && data[1] == "U" {
			if data[3] == user && data[0] != "0" {
				return &authentication.User{
					Username: data[3],
					Password: data[4],
					UID:      data[0],
					Group:    data[2],
				}, i
			}
		}

	}

	return nil, -1
}
