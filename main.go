package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	name    string
	command string
}

func (i item) Title() string       { return i.name }
func (i item) Description() string { return i.command }
func (i item) FilterValue() string { return i.name }

type model struct {
	list     list.Model
	selected string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			m.selected = m.list.SelectedItem().(item).name
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

type PackageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

func main() {
	// Open the package.json file
	file, err := os.Open("package.json")
	if err != nil {
		fmt.Println("No package.json found in the current directory")
		os.Exit(0)
	}
	defer file.Close()

	// Detect package manager
	packageManager := "npm"
	if _, err := os.Stat("package-lock.json"); err == nil {
		packageManager = "npm"
	} else if _, err := os.Stat("yarn.lock"); err == nil {
		packageManager = "yarn"
	} else if _, err := os.Stat("pnpm-lock.yaml"); err == nil {
		packageManager = "pnpm"
	} else if _, err := os.Stat("bun.lockb"); err == nil {
		packageManager = "bun"
	}

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

	items := []list.Item{}
	for scriptName, scriptCommand := range packageData.Scripts {
		items = append(items, item{
			name:    scriptName,
			command: scriptCommand,
		})
	}

	// Sort the items alphabetically by command name
	sort.Slice(items, func(i, j int) bool {
		return items[i].(item).name < items[j].(item).name
	})

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "Please choose a script to run"

	p := tea.NewProgram(m, tea.WithAltScreen())

	final, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	// Assert the final tea.Model to our local model and print the choice.
	if m, ok := final.(model); ok && m.selected != "" {
		fmt.Printf("\n---\nExecuting %s run %s\n", packageManager, m.selected)

		// Define the command to run
		cmd := exec.Command(packageManager, "run", m.selected)

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
