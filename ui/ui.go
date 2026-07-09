package ui

import (
	"fmt"
	"io"
	"strings"

	"ddev-clim/config"
	"ddev-clim/ddev"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "15", Dark: "15"}). // White text on header
			Background(lipgloss.Color("4")).                            // Standard Blue
			Padding(0, 1)
	statusRunning    = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Standard Green
	statusStopped    = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Bright Black (Gray)
	statusWorking    = lipgloss.NewStyle().Foreground(lipgloss.Color("5")) // Standard Magenta
	statusUnhealthy  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Standard Red
	instructionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
	pathStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	headerStyle      = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "15"}).Bold(true).Underline(true)
	fuchsia          = lipgloss.Color("5") // Standard Magenta
)

type item struct {
	project    ddev.Project
	processing bool
	spinner    string // Current spinner frame
	lastError  string
}

func (i item) Title() string       { return i.project.Name }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.project.Name }

type itemDelegate struct {
	spinner spinner.Model
}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := ""
	
	// Column widths
	nameWidth := 30
	statusWidth := 15

	// Name
	name := i.project.Name
	if runewidth.StringWidth(name) > nameWidth-2 {
		name = runewidth.Truncate(name, nameWidth-5, "...")
	}
	
	// Default foreground for normal titles (adapts to light/dark automatically)
	nameStyle := lipgloss.NewStyle().Bold(true)
	if index == m.Index() {
		nameStyle = nameStyle.Foreground(fuchsia)
	}
	
	nameStr := runewidth.FillRight(name, nameWidth)
	str += nameStyle.Render(nameStr)

	// Status
	status := i.project.Status
	var statusStr string
	if i.processing {
		action := "starting..."
		if status == "running" || status == "OK" || status == "unhealthy" {
			action = "stopping..."
		}
		statusStr = statusWorking.Render(i.spinner + " " + action)
	} else if i.lastError != "" {
		statusStr = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("failed")
	} else if status == "running" || status == "OK" {
		statusStr = statusRunning.Render(status)
	} else if status == "unhealthy" {
		statusStr = statusUnhealthy.Render(status)
	} else {
		statusStr = statusStopped.Render(status)
	}
	
	// Manual padding for status
	rawStatus := status
	if i.processing {
		if status == "running" || status == "OK" || status == "unhealthy" {
			rawStatus = "stopping..."
		} else {
			rawStatus = "starting..."
		}
	} else if i.lastError != "" {
		rawStatus = "failed"
	}
	padding := statusWidth - runewidth.StringWidth(rawStatus)
	if padding < 0 {
		padding = 0
	}
	str += statusStr + strings.Repeat(" ", padding)

	// URL
	url := i.project.PrimaryURL
	str += url

	// Selection indicator (left border replacement)
	if index == m.Index() {
		prefix := lipgloss.NewStyle().Foreground(fuchsia).Render("> ")
		fmt.Fprintf(w, "%s%s", prefix, str)
	} else {
		fmt.Fprintf(w, "  %s", str)
	}
}

type statusMsg struct {
	index      int
	isStopping bool
	err        error
}

type refreshMsg struct {
	projects []ddev.Project
	err      error
}

type model struct {
	list           list.Model
	config         *config.Config
	spinner        spinner.Model
	isRefreshing   bool
	refreshMsg     string
	err            error
	lastErrors     map[string]string
	terminalWidth  int
	terminalHeight int
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.refreshCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.list.FilterState() != list.Filtering {
			if msg.String() == "enter" || msg.String() == " " {
				idx := m.list.Index()
				items := m.list.Items()
				if idx < 0 || idx >= len(items) {
					return m, nil
				}
				i, ok := items[idx].(item)
				if ok && !i.processing {
					i.processing = true
					i.spinner = m.spinner.View()
					m.list.SetItem(idx, i)

					isStopping := i.project.Status == "running" || i.project.Status == "OK" || i.project.Status == "unhealthy"

					return m, func() tea.Msg {
						var err error
						if isStopping {
							err = ddev.StopProject(i.project.AppRoot)
						} else {
							err = ddev.StartProject(i.project.AppRoot)
						}
						return statusMsg{index: idx, isStopping: isStopping, err: err}
					}
				}
			}
			if msg.String() == "p" {
				m.isRefreshing = true
				m.refreshMsg = "Powering off DDEV..."
				return m, func() tea.Msg {
					err := ddev.Poweroff()
					return statusMsg{index: -1, err: err}
				}
			}
		}
	case statusMsg:
		if msg.index >= 0 {
			items := m.list.Items()
			if msg.index < len(items) {
				if i, ok := items[msg.index].(item); ok {
					name := i.project.Name
					if msg.err != nil {
						m.lastErrors[name] = msg.err.Error()
					} else {
						delete(m.lastErrors, name)
						// Update config only on successful start/stop
						if msg.isStopping {
							m.config.RemoveProject(name)
						} else {
							m.config.AddProject(name)
						}
						_ = config.SaveConfig(m.config)
					}
				}
			}
		} else {
			if msg.err != nil {
				m.err = msg.err
			} else {
				m.err = nil
				m.lastErrors = make(map[string]string)
				// Clear all running projects from config on global poweroff
				m.config.RunningProjects = []string{}
				_ = config.SaveConfig(m.config)
			}
		}
		m.refreshMsg = "Loading DDEV projects..."
		return m, m.refreshCmd()

	case refreshMsg:
		m.isRefreshing = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		items := []list.Item{}
		for _, p := range msg.projects {
			errStr := m.lastErrors[p.Name]
			items = append(items, item{project: p, lastError: errStr})
		}
		m.list.SetItems(items)
		m = m.updateListSize()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		
		// Update processing items with new spinner frame
		items := m.list.Items()
		for idx, listItem := range items {
			if i, ok := listItem.(item); ok && i.processing {
				i.spinner = m.spinner.View()
				// Also update the delegate's spinner reference if needed
				m.list.SetItem(idx, i)
			}
		}
		return m, cmd

	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		m = m.updateListSize()
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) updateListSize() model {
	if m.terminalHeight == 0 {
		return m
	}
	h, v := docStyle.GetFrameSize()
	
	// 7 lines for headers, title, help and spacing
	desiredHeight := len(m.list.Items()) + 7
	
	maxHeight := m.terminalHeight - v - 6
	if desiredHeight > maxHeight {
		desiredHeight = maxHeight
	}
	if desiredHeight < 5 {
		desiredHeight = 5
	}
	
	m.list.SetSize(m.terminalWidth-h, desiredHeight)
	return m
}

func (m model) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		projects, err := ddev.GetProjects()
		return refreshMsg{projects: projects, err: err}
	}
}

func (m model) View() string {
	if m.isRefreshing {
		msg := "Loading DDEV projects..."
		if m.refreshMsg != "" {
			msg = m.refreshMsg
		}
		return docStyle.Render(fmt.Sprintf("\n\n  %s %s", m.spinner.View(), msg))
	}

	s := docStyle.Render(m.list.View())
	
	// Centralized Details Panel
	var details string
	idx := m.list.Index()
	items := m.list.Items()
	if idx >= 0 && idx < len(items) {
		if i, ok := items[idx].(item); ok {
			p := i.project
			
			borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
			boldLabel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
			
			details = "\n" + borderStyle.Render("── SELECTED PROJECT DETAILS ──────────────────────────────────────────") + "\n"
			details += fmt.Sprintf("%s %-12s %s %-10s %s %s\n", 
				boldLabel.Render("Name:"), p.Name,
				boldLabel.Render("Type:"), p.Type,
				boldLabel.Render("Path:"), p.AppRoot,
			)
			
			// Show errors or status info
			lastError := m.lastErrors[p.Name]
			if lastError != "" {
				errLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render("❌ Error:")
				formattedErr := formatError(lastError, 3)
				errLines := strings.Split(formattedErr, "\n")
				details += fmt.Sprintf("%s %s\n", errLabel, errLines[0])
				for _, line := range errLines[1:] {
					details += fmt.Sprintf("         %s\n", line)
				}
			} else if p.Status == "unhealthy" {
				tipLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true).Render("⚠️ Unhealthy:")
				details += fmt.Sprintf("%s Project is unhealthy. Try toggling it to stop/restart, or press 'p' to poweroff DDEV globally.\n", tipLabel)
			} else if p.Status == "running" || p.Status == "OK" {
				okLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true).Render("✓ Running:")
				details += fmt.Sprintf("%s at %s\n", okLabel, p.PrimaryURL)
			} else {
				stoppedLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true).Render("○ Stopped:")
				details += fmt.Sprintf("%s Project is stopped. Toggle (enter/space) to start it.\n", stoppedLabel)
			}
		}
	}
	
	if details != "" {
		s += details
	}

	if m.err != nil {
		s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("Global Error: %v", m.err))
	}
	
	// Lock shortcuts to the bottom of the terminal
	shortcuts := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
		"  ↑/k: up/down • /: filter • enter/space: toggle • p: poweroff • q: quit",
	)
	
	if m.terminalHeight > 0 {
		lines := strings.Split(s, "\n")
		currentLines := len(lines)
		targetLines := m.terminalHeight - 2
		if currentLines < targetLines {
			s += strings.Repeat("\n", targetLines - currentLines)
		}
	}
	s += "\n" + shortcuts
	
	return s
}

func formatError(errStr string, maxLines int) string {
	lines := strings.Split(errStr, "\n")
	var nonEnvLines []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			nonEnvLines = append(nonEnvLines, l)
		}
	}
	
	if len(nonEnvLines) <= maxLines {
		return strings.Join(nonEnvLines, "\n")
	}
	
	// Get the last maxLines
	return "... (truncated) ...\n" + strings.Join(nonEnvLines[len(nonEnvLines)-maxLines:], "\n")
}

func StartTUI() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(fuchsia)

	m := model{
		config:       cfg,
		spinner:      sp,
		isRefreshing: true,
		refreshMsg:   "Loading DDEV projects...",
		lastErrors:   make(map[string]string),
	}

	d := itemDelegate{spinner: sp}

	l := list.New([]list.Item{}, d, 0, 0)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle() // Unset default title styling
	
	instr := "\n" + instructionStyle.Render("Navigate the instances, toggle with enter/space, or poweroff with p")
	
	// Pre-header padding
	nameWidth := 30
	statusWidth := 15
	header := "  " + 
		runewidth.FillRight("NAME", nameWidth) +
		runewidth.FillRight("STATUS", statusWidth) +
		"URL"
	
	l.Title = titleStyle.Render("DDEV CLInstance Manager") + instr + "\n\n" + headerStyle.Render(header)
	
	m.list = l

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
