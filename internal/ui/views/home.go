package views

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type job struct {
	title    string
	team     string
	location string
	body     string
}

var jobs = []job{
	{
		title:    "Go engineer, session server",
		team:     "Infrastructure",
		location: "Remote, Europe",
		body: `You will own the server that keeps terminal sessions alive
when the client goes away. Go, Linux, a lot of file descriptors.

We care about latency more than features.`,
	},
	{
		title:    "Terminal rendering engineer",
		team:     "Client",
		location: "Remote",
		body: `You will work on the part that draws cells on screen: scrollback,
selection, resize, Unicode width. Experience with a terminal emulator
codebase is a big plus.`,
	},
	{
		title:    "Product designer",
		team:     "Design",
		location: "Remote",
		body: `Developer tools, keyboard-first interfaces, a lot of monospace.
You will design for the terminal, the web and macOS at the same time.`,
	},
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	metaStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	okStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
)

type board struct {
	cursor  int
	open    bool
	applied bool
	visitor string
	width   int
	height  int
}

type appliedMsg struct{ err error }

func apply(visitor string, j job) tea.Cmd {
	return func() tea.Msg {
		f, err := os.OpenFile("applications.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return appliedMsg{err}
		}
		defer f.Close()
		line := fmt.Sprintf("%s\t%s\t%s\n", time.Now().Format(time.RFC3339), visitor, j.title)
		_, err = f.WriteString(line)
		return appliedMsg{err}
	}
}

func (b board) Init() tea.Cmd {
	return nil
}

func (b board) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
		b.height = msg.Height
	case appliedMsg:
		b.applied = msg.err == nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return b, tea.Quit
		case "up", "k":
			if !b.open && b.cursor > 0 {
				b.cursor--
			}
		case "down", "j":
			if !b.open && b.cursor < len(jobs)-1 {
				b.cursor++
			}
		case "enter", "right", "l":
			b.open = true
		case "esc", "left", "h":
			b.open = false
		case "a":
			if b.open && b.visitor != "anonymous" {
				return b, apply(b.visitor, jobs[b.cursor])
			}
		}
	}
	return b, nil
}

func (b board) View() tea.View {
	var s strings.Builder
	s.WriteString(titleStyle.Render("Jobs at Zander") + "\n\n")

	if b.open {
		j := jobs[b.cursor]
		s.WriteString(titleStyle.Render(j.title) + "\n")
		s.WriteString(metaStyle.Render(j.team+" • "+j.location) + "\n\n")
		s.WriteString(j.body + "\n\n")
		switch {
		case b.applied:
			s.WriteString(okStyle.Render("Saved. We know your key, we will be in touch.") + "\n\n")
		case b.visitor == "anonymous":
			s.WriteString(metaStyle.Render("Connect with an SSH key to apply from here.") + "\n\n")
		default:
			s.WriteString(helpStyle.Render("a apply · "))
		}
		s.WriteString(helpStyle.Render("esc back • q quit"))
	} else {
		for i, j := range jobs {
			cursor := " "
			if i == b.cursor {
				cursor = cursorStyle.Render("> ")
			}
			s.WriteString(cursor + j.title + "\n")
			s.WriteString("  " + metaStyle.Render(j.team+" • "+j.location) + "\n\n")
		}
		s.WriteString(helpStyle.Render("↑/↓ move · enter open · q quit"))
	}

	s.WriteString("\n\n" + metaStyle.Render("you: "+b.visitor))

	v := tea.NewView(s.String())
	v.AltScreen = true
	return v
}

func NewBoard(visitor string, width, height int) tea.Model {
	return board{visitor: visitor, width: width, height: height}
}
