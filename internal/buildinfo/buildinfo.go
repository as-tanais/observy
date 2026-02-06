// buildinfo - пакет для отображения информации при старте приложения
package buildinfo

import "fmt"

// PrintInfo - При старте приложения выводите в stdout сообщение в следующем формате
func PrintInfo(version, date, commit string) {

	if version == "" {
		version = "N/A"
	}

	if version == "" {
		version = "N/A"
	}

	if version == "" {
		version = "N/A"
	}

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}
