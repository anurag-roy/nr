package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type PackageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

func main() {
	// Open the package.json file
	file, err := os.Open("package.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Read the file content into a byte slice
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Create a variable to hold the parsed JSON data
	var packageData PackageJSON

	// Unmarshal the JSON data into the PackageJSON struct
	err = json.Unmarshal(data, &packageData)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	// Print the extracted scripts
	fmt.Println("Scripts:")
	for scriptName, scriptCommand := range packageData.Scripts {
		fmt.Printf("%s: %s\n", scriptName, scriptCommand)
	}
}
