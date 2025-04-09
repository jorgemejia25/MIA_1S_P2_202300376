package utils

import (
	"strings"

	"disk.simulator.com/m/v2/internal/disk/types/structures/authentication"
)

func FindLastGroupInFile(
	fileText string,
) authentication.Group {
	lines := strings.Split(fileText, "\n")

	var lastGroup authentication.Group

	for _, line := range lines {
		// split line by comma
		data := strings.Split(line, ",")

		if len(data) > 1 && data[1] == "G" {
			lastGroup = authentication.Group{
				GID:   data[0],
				Name:  data[2],
			}
		}
	}

	return lastGroup
}
