package utils

import (
	"strings"

	"disk.simulator.com/m/v2/internal/disk/types/structures/authentication"
)

func FindGroupInFile(fileText string, grp string) (*authentication.Group, int) {
	lines := strings.Split(fileText, "\n")

	for i, line := range lines {
		data := strings.Split(line, ",")
		if len(data) > 1 && data[1] == "G" && data[2] == grp && data[0] != "0" {
			return &authentication.Group{
				GID:  data[0],
				Name: data[2],
			}, i
		}
	}
	return nil, -1
}
