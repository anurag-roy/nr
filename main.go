package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type PackageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

type script struct {
	name    string
	command string
}

type model struct {
	choices []script
	cursor  int
	choice  string
}

func main() {
	// Open the package.json file
	file, err := os.Open("package.json")
	if err != nil {
		fmt.Println("No package.json found in the current directory")
		os.Exit(0)
	}
	defer file.Close()

	// Read the file content into a byte slice
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading package.json file:", err)
		os.Exit(1)
	}

	// Create a variable to hold the parsed JSON data
	var packageData PackageJSON

	// Unmarshal the JSON data into the PackageJSON struct
	if err = json.Unmarshal(data, &packageData); err != nil {
		fmt.Println("Not a valid JSON file")
		os.Exit(1)
	}

	scripts := []script{}
	for scriptName, scriptCommand := range packageData.Scripts {
		scripts = append(scripts, script{
			name:    scriptName,
			command: scriptCommand,
		})
	}

	p := tea.NewProgram(model{
		choices: scripts,
	})

	m, err := p.Run()
	if err != nil {
		fmt.Println("Failed to start program:", err)
		os.Exit(1)
	}

	// Assert the final tea.Model to our local model and print the choice.
	if m, ok := m.(model); ok && m.choice != "" {
		fmt.Printf("\n---\nExecuting npm run %s!\n", m.choice)

		// Define the command to run
		cmd := exec.Command("npm", "run", m.choice)

		// Get a pipe to capture command's standard output
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Println("Error capturing stdout:", err)
			os.Exit(1)
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			fmt.Println("Error starting the command:", err)
			os.Exit(1)
		}

		// Read and print the output line by line
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}

		// Wait for the command to finish
		if err := cmd.Wait(); err != nil {
			fmt.Println("Error while waiting for the command to finish:", err)
			os.Exit(1)
		}
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "enter":
			// Send the choice on the channel and exit.
			m.choice = m.choices[m.cursor].name
			return m, tea.Quit

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	s := strings.Builder{}
	s.WriteString("Which script would you like to run?\n\n")

	for i := 0; i < len(m.choices); i++ {
		if m.cursor == i {
			s.WriteString("[•] ")
		} else {
			s.WriteString("[ ] ")
		}
		s.WriteString(m.choices[i].name)
		s.WriteString("\n")
	}
	s.WriteString("\n(press q to quit)\n")

	return s.String()
}
