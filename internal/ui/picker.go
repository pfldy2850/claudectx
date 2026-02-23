package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PickerItem represents a context in the interactive picker.
type PickerItem struct {
	Name        string
	Description string
	Files       int
	TotalSize   int64
	UpdatedAt   time.Time
	IsCurrent   bool
}

type pickerModel struct {
	items    []PickerItem
	cursor   int
	selected string
	quitting bool
}

func (m pickerModel) Init() tea.Cmd {
	return nil
}

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.items[m.cursor].Name
			return m, tea.Quit
		}
	}
	return m, nil
}

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	currentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
)

func (m pickerModel) View() string {
	if m.quitting && m.selected == "" {
		return ""
	}

	s := titleStyle.Render("Select a context:") + "\n\n"

	for i, item := range m.items {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "> "
			style = selectedStyle
		}

		name := item.Name
		if item.IsCurrent {
			name += currentStyle.Render(" (active)")
		}

		line := fmt.Sprintf("%s%s", cursor, style.Render(name))
		detail := dimStyle.Render(fmt.Sprintf("    %d files, %s, updated %s",
			item.Files, formatSizeUI(item.TotalSize), item.UpdatedAt.Format("2006-01-02")))

		if item.Description != "" {
			detail += dimStyle.Render(fmt.Sprintf("\n    %s", item.Description))
		}

		s += line + "\n" + detail + "\n\n"
	}

	s += dimStyle.Render("↑/↓ navigate • enter select • q quit")

	return s
}

// RunPicker displays the interactive context picker and returns the selected context name.
func RunPicker(items []PickerItem) (string, error) {
	if len(items) == 0 {
		return "", nil
	}

	m := pickerModel{items: items}
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return "", err
	}

	final := result.(pickerModel)
	return final.selected, nil
}

func formatSizeUI(bytes int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
	)
	switch {
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
