package utils



// PrintDirectories imprime un slice de directorios para diagnóstico
func PrintDirectories(dirs []string) string {
	result := "["
	for i, dir := range dirs {
		if i > 0 {
			result += ", "
		}
		result += "\"" + dir + "\""
	}
	result += "]"
	return result
}
