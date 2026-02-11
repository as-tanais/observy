package buildinfo

import "fmt"

// PrintInfo печатает информацию о билде в stdout
func PrintInfo(version, date, commit string) {
	version = normalizeString(version, "N/A")
	date = normalizeString(date, "N/A")
	commit = normalizeString(commit, "N/A")

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}

// normalizeString возвращает значение или значение по умолчанию, если исходное пустое
func normalizeString(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
