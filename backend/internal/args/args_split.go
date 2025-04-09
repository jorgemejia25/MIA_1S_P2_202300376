package args

import (
	"regexp"
	"strings"
)

// SplitArgs divide una cadena en argumentos, respetando las comillas y los flags con valores unidos por "=".
func SplitArgs(command string) []string {
	re := regexp.MustCompile(`(-{1,2}[^=\s]+)(=([^"\s]+|"[^"]+"|'[^']+'))?|[^\s"']+|"([^"]*)"|'([^']*)'`)
	matches := re.FindAllStringSubmatch(command, -1)

	result := make([]string, 0, len(matches))

	for _, match := range matches {
		if match[1] != "" {
			// Convertir -flag a --flag y normalizar a minúsculas
			flag := match[1]
			if len(flag) > 2 && flag[0] == '-' && flag[1] != '-' {
				flag = "-" + flag
			}
			// Normalizar flags a minúsculas
			flag = strings.ToLower(flag)
			result = append(result, flag)

			if match[2] != "" {
				value := strings.Trim(match[3], `"'`)
				result = append(result, value)
			}
		} else if match[3] != "" || match[4] != "" {
			value := strings.Trim(match[3]+match[4], `"'`)
			result = append(result, value)
		} else {
			result = append(result, match[0])
		}
	}

	return result
}
