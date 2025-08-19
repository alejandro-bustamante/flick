package tui

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item string

func (i item) Title() string       { return string(i) }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return string(i) }

type model struct {
	currentDir string
	items      []list.Item
	list       list.Model
	err        error
}

var (
	leftPaneWidth  = 40
	rightPaneWidth = 80
	styles         = struct {
		leftPane  lipgloss.Style
		rightPane lipgloss.Style
	}{
		leftPane: lipgloss.NewStyle().
			Width(leftPaneWidth).
			Border(lipgloss.RoundedBorder()).
			BorderRight(true).
			Padding(0, 1),

		rightPane: lipgloss.NewStyle().
			Width(rightPaneWidth).
			Padding(0, 1),
	}
)

func initialModel() model {
	dir, _ := os.Getwd()
	l := list.New(nil, list.NewDefaultDelegate(), leftPaneWidth-4, 20)
	l.Title = "Explorador de Archivos"
	return model{
		currentDir: dir,
		list:       l,
	}
}

func (m model) Init() tea.Cmd {
	return m.readDir()
}

func (m model) readDir() tea.Cmd {
	return func() tea.Msg {
		files, err := ioutil.ReadDir(m.currentDir)
		if err != nil {
			return errMsg{err}
		}

		var items []list.Item
		// Agrega opción para retroceder
		if parent := filepath.Dir(m.currentDir); parent != m.currentDir {
			items = append(items, item(".."))
		}

		for _, f := range files {
			name := f.Name()
			if f.IsDir() {
				name += "/"
			}
			items = append(items, item(name))
		}
		return itemsMsg(items)
	}
}

type itemsMsg []list.Item
type errMsg struct{ error }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case itemsMsg:
		m.items = msg
		m.list.SetItems(msg)
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit

		case "enter":
			selected := m.list.SelectedItem().(item)
			selectedStr := string(selected)
			if selectedStr == ".." {
				m.currentDir = filepath.Dir(m.currentDir)
			} else {
				path := filepath.Join(m.currentDir, strings.TrimSuffix(selectedStr, "/"))
				info, err := os.Stat(path)
				if err == nil && info.IsDir() {
					m.currentDir = path
				}
			}
			return m, m.readDir()
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	left := styles.leftPane.Render(m.list.View())
	right := styles.rightPane.Render("") // Reservado para futuro
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func main() {
	if err := tea.NewProgram(initialModel()).Start(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
