package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

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
	latestLog  string // Live stream output
}

func (i item) Title() string       { return i.project.Name }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.project.Name }

// globalItem is the synthetic "DDEV Global" row always at position 0.
type globalItem struct{}

func (g globalItem) Title() string       { return "ddev" }
func (g globalItem) Description() string { return "" }
func (g globalItem) FilterValue() string { return "ddev" }

type itemDelegate struct {
	spinner spinner.Model
}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	// Column widths (shared)
	nameWidth := 30
	statusWidth := 15

	// ── Global row ───────────────────────────────────────────────────────────
	if _, ok := listItem.(globalItem); ok {
		dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		selected := index == m.Index()

		nameStr := runewidth.FillRight("ddev", nameWidth)
		var nameRendered string
		if selected {
			nameRendered = lipgloss.NewStyle().Foreground(fuchsia).Render(nameStr)
		} else {
			nameRendered = dimmed.Render(nameStr)
		}

		// Count running projects from the list
		running := 0
		for _, li := range m.Items() {
			if proj, ok := li.(item); ok {
				if proj.project.Status == "running" || proj.project.Status == "OK" {
					running++
				}
			}
		}
		var statusRendered string
		if running > 0 {
			statusRendered = dimmed.Render(fmt.Sprintf("%d running", running))
		} else {
			statusRendered = dimmed.Render("idle")
		}

		prefix := "  "
		if selected {
			prefix = lipgloss.NewStyle().Foreground(fuchsia).Render("> ")
		}
		fmt.Fprintf(w, "%s%s%s", prefix, nameRendered, statusRendered)
		return
	}

	// ── Project row ──────────────────────────────────────────────────────────
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := ""

	// Name
	name := i.project.Name
	if runewidth.StringWidth(name) > nameWidth-2 {
		name = runewidth.Truncate(name, nameWidth-5, "...")
	}
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
	str += i.project.PrimaryURL

	// Selection indicator
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

type describeMsg struct {
	projectName string
	raw         ddev.DescribeRaw
	err         error
}

type pollDescribeMsg struct {
	projectName string
	appRoot     string
}

type streamLineMsg struct {
	projectName string
	isStopping  bool
	line        string
	stream      *ddev.CommandStream
}

type streamFinishedMsg struct {
	projectName string
	isStopping  bool
	err         error
}

type model struct {
	list               list.Model
	config             *config.Config
	spinner            spinner.Model
	isRefreshing       bool
	refreshMsg         string
	err                error
	lastErrors         map[string]string
	terminalWidth      int
	terminalHeight     int
	showPoweroffPrompt bool
	descriptions       map[string]ddev.DescribeRaw
	latestLogs         map[string]string
	accumulatedLogs    map[string][]string
	isPoweringOff      bool
	poweroffLogs       []string
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
		
		if m.showPoweroffPrompt {
			if msg.String() == "y" || msg.String() == "Y" {
				m.showPoweroffPrompt = false
				stream, err := ddev.StartCommandStream(".", "poweroff")
				if err != nil {
					m.err = err
					return m, nil
				}
				m.isPoweringOff = true
				m.poweroffLogs = []string{"Initiating poweroff..."}
				// Jump cursor to the global row (always index 0)
				m.list.Select(0)
				return m, readStreamLineCmd("GLOBAL_POWEROFF", false, stream)
			}
			if msg.String() == "n" || msg.String() == "N" || msg.String() == "esc" {
				m.showPoweroffPrompt = false
				return m, nil
			}
			return m, nil // Block other keys while prompt is active
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
					i.latestLog = "starting..."
					m.list.SetItem(idx, i)

					isStopping := i.project.Status == "running" || i.project.Status == "OK" || i.project.Status == "unhealthy"
					if isStopping {
						i.latestLog = "stopping..."
						m.list.SetItem(idx, i)
					}

					action := "start"
					if isStopping {
						action = "stop"
					}
					stream, err := ddev.StartCommandStream(i.project.AppRoot, action)
					if err != nil {
						i.processing = false
						i.lastError = err.Error()
						m.list.SetItem(idx, i)
						return m, nil
					}

					m.latestLogs[i.project.Name] = i.latestLog
					m.accumulatedLogs[i.project.Name] = []string{i.latestLog}

					return m, tea.Batch(
						readStreamLineCmd(i.project.Name, isStopping, stream),
						m.pollDescribeCmd(i.project.Name, i.project.AppRoot),
					)
				}
			}
			if msg.String() == "p" {
				stream, err := ddev.StartCommandStream(".", "poweroff")
				if err != nil {
					m.err = err
					return m, nil
				}
				m.isPoweringOff = true
				m.poweroffLogs = []string{"Initiating poweroff..."}
				// Jump cursor to the global row (always index 0)
				m.list.Select(0)
				return m, readStreamLineCmd("GLOBAL_POWEROFF", false, stream)
			}
		}
	case statusMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.err = nil
			m.lastErrors = make(map[string]string)
			// Clear all running projects from config on global poweroff
			m.config.RunningProjects = []string{}
			_ = config.SaveConfig(m.config)
		}
		m.refreshMsg = "Loading DDEV projects..."
		return m, m.refreshCmd()

	case streamLineMsg:
		if msg.projectName == "GLOBAL_POWEROFF" {
			m.poweroffLogs = append(m.poweroffLogs, msg.line)
			return m, readStreamLineCmd(msg.projectName, msg.isStopping, msg.stream)
		}

		m.latestLogs[msg.projectName] = msg.line
		m.accumulatedLogs[msg.projectName] = append(m.accumulatedLogs[msg.projectName], msg.line)
		
		idx := -1
		items := m.list.Items()
		for i, listItem := range items {
			if item, ok := listItem.(item); ok && item.project.Name == msg.projectName {
				idx = i
				break
			}
		}
		if idx >= 0 {
			if item, ok := items[idx].(item); ok {
				item.spinner = m.spinner.View()
				item.latestLog = msg.line
				m.list.SetItem(idx, item)
			}
		}
		return m, readStreamLineCmd(msg.projectName, msg.isStopping, msg.stream)

	case streamFinishedMsg:
		if msg.projectName == "GLOBAL_POWEROFF" {
			m.isPoweringOff = false
			m.poweroffLogs = nil
			if msg.err != nil {
				m.err = msg.err
			} else {
				m.err = nil
				m.lastErrors = make(map[string]string)
				m.config.RunningProjects = []string{}
				_ = config.SaveConfig(m.config)
			}
			m.refreshMsg = "Loading DDEV projects..."
			return m, m.refreshCmd()
		}

		idx := -1
		items := m.list.Items()
		for i, listItem := range items {
			if item, ok := listItem.(item); ok && item.project.Name == msg.projectName {
				idx = i
				break
			}
		}
		
		if idx >= 0 {
			if item, ok := items[idx].(item); ok {
				item.processing = false
				item.latestLog = ""
				m.list.SetItem(idx, item)
			}
		}
		
		if msg.err != nil {
			m.lastErrors[msg.projectName] = strings.Join(m.accumulatedLogs[msg.projectName], "\n")
			m.showPoweroffPrompt = true
		} else {
			delete(m.lastErrors, msg.projectName)
			if msg.isStopping {
				m.config.RemoveProject(msg.projectName)
			} else {
				m.config.AddProject(msg.projectName)
			}
			_ = config.SaveConfig(m.config)
		}
		
		delete(m.latestLogs, msg.projectName)
		delete(m.accumulatedLogs, msg.projectName)
		
		return m, m.refreshCmd()

	case pollDescribeMsg:
		items := m.list.Items()
		stillProcessing := false
		for _, listItem := range items {
			if item, ok := listItem.(item); ok && item.project.Name == msg.projectName && item.processing {
				stillProcessing = true
				break
			}
		}
		if stillProcessing {
			return m, tea.Batch(
				m.describeCmd(msg.projectName, msg.appRoot),
				m.pollDescribeCmd(msg.projectName, msg.appRoot),
			)
		}
		return m, nil

	case refreshMsg:
		m.isRefreshing = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		// Always prepend the global row
		items := []list.Item{globalItem{}}
		for _, p := range msg.projects {
			errStr := m.lastErrors[p.Name]
			items = append(items, item{project: p, lastError: errStr})
		}
		m.list.SetItems(items)
		m = m.updateListSize()
		
		// Query describe for currently selected project (skip global row at 0)
		idx := m.list.Index()
		if idx > 0 && idx < len(items) {
			if i, ok := items[idx].(item); ok {
				return m, m.describeCmd(i.project.Name, i.project.AppRoot)
			}
		}
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
	oldIndex := m.list.Index()
	m.list, cmd = m.list.Update(msg)
	
	newIndex := m.list.Index()
	if oldIndex != newIndex && newIndex > 0 {
		items := m.list.Items()
		if newIndex < len(items) {
			if i, ok := items[newIndex].(item); ok {
				describeCmd := m.describeCmd(i.project.Name, i.project.AppRoot)
				return m, tea.Batch(cmd, describeCmd)
			}
		}
	}
	return m, cmd
}

func (m model) updateListSize() model {
	if m.terminalHeight == 0 {
		return m
	}
	h, v := docStyle.GetFrameSize()
	
	// 7 lines for headers, title, help and spacing
	desiredHeight := len(m.list.Items()) + 7
	
	maxHeight := m.terminalHeight - v - 7
	if desiredHeight > maxHeight {
		desiredHeight = maxHeight
	}
	if desiredHeight < 5 {
		desiredHeight = 5
	}
	
	m.list.SetSize(m.terminalWidth-h, desiredHeight)
	return m
}

func (m model) describeCmd(name string, appRoot string) tea.Cmd {
	return func() tea.Msg {
		raw, err := ddev.DescribeProject(appRoot)
		return describeMsg{projectName: name, raw: raw, err: err}
	}
}

func (m model) pollDescribeCmd(name string, appRoot string) tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return pollDescribeMsg{projectName: name, appRoot: appRoot}
	})
}

func readStreamLineCmd(name string, isStopping bool, stream *ddev.CommandStream) tea.Cmd {
	return func() tea.Msg {
		buf := make([]byte, 1024)
		n, err := stream.Reader.Read(buf)
		if err != nil {
			exitErr := <-stream.WaitCh
			// Put it back in the channel in case other reads also hit EOF
			stream.WaitCh <- exitErr
			return streamFinishedMsg{
				projectName: name,
				isStopping:  isStopping,
				err:         exitErr,
			}
		}
		
		chunk := string(buf[:n])
		parts := strings.FieldsFunc(chunk, func(r rune) bool {
			return r == '\n' || r == '\r'
		})
		
		var line string
		if len(parts) > 0 {
			line = strings.TrimSpace(parts[len(parts)-1])
		}
		
		if line == "" {
			line = strings.TrimSpace(chunk)
		}
		
		return streamLineMsg{
			projectName: name,
			isStopping:  isStopping,
			line:        line,
			stream:      stream,
		}
	}
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
	if m.showPoweroffPrompt {
		borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
		boldLabel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208"))
		
		details = "\n" + borderStyle.Render("── RECOMMENDATION ────────────────────────────────────────────────────") + "\n"
		details += fmt.Sprintf("%s DDEV action failed. We recommend running poweroff to reset container networking.\n", boldLabel.Render("Recommendation:"))
		details += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Render("   Power off DDEV globally now? (y/n) ")
	} else {
		idx := m.list.Index()
		items := m.list.Items()
		if idx == 0 {
			// Global row selected
			borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			boldLabel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
			details = "\n" + borderStyle.Render("── DDEV GLOBAL STATUS ────────────────────────────────────────────────") + "\n"
			if m.isPoweringOff {
				details += fmt.Sprintf("%s Powering off DDEV globally...\n", m.spinner.View())
				logs := m.poweroffLogs
				start := len(logs) - 3
				if start < 0 {
					start = 0
				}
				if len(logs) > 0 {
					rule := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("-", 40))
					details += fmt.Sprintf("%s\n%s\n", boldLabel.Render("Progress:"), rule)
					for i := start; i < len(logs); i++ {
						details += fmt.Sprintf("%s\n", logs[i])
					}
				}
			} else {
				dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				details += dimmed.Render("  Press p to power off all DDEV projects and router globally.") + "\n"
			}
		} else if idx > 0 && idx < len(items) {
			if i, ok := items[idx].(item); ok {
				p := i.project
				
				borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
				boldLabel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
				
				details = "\n" + borderStyle.Render("── PROJECT STATUS CENTER ─────────────────────────────────────────────") + "\n"
				
				desc, hasDesc := m.descriptions[p.Name]
				if hasDesc && (p.Status != "stopped" || i.processing) {
					typeStr := p.Type
					if desc.WebserverType != "" {
						typeStr = fmt.Sprintf("%s (%s)", p.Type, desc.WebserverType)
					}
					
					details += fmt.Sprintf("%s %-12s %s %-14s %s %-8s %s %s\n", 
						boldLabel.Render("Name:"), p.Name,
						boldLabel.Render("Type:"), typeStr,
						boldLabel.Render("PHP:"), desc.PhpVersion,
						boldLabel.Render("DB:"), fmt.Sprintf("%s:%s", desc.DatabaseType, desc.DatabaseVersion),
					)
					
					statusOKStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
					statusWarnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
					statusErrorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
					statusOffStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
					
					formatServiceStatus := func(svcName, status string) string {
						var styled string
						if status == "running" || status == "OK" {
							styled = statusOKStyle.Render(status)
						} else if status == "unhealthy" {
							styled = statusErrorStyle.Render(status)
						} else if status == "starting" {
							styled = statusWarnStyle.Render(status)
						} else {
							styled = statusOffStyle.Render(status)
						}
						return fmt.Sprintf("%s: %s", svcName, styled)
					}
					
					webStatus := "stopped"
					if web, ok := desc.Services["web"]; ok {
						webStatus = web.Status
					}
					dbStatus := "stopped"
					if db, ok := desc.Services["db"]; ok {
						dbStatus = db.Status
					}
					
					containerStatuses := []string{
						formatServiceStatus("web", webStatus),
						formatServiceStatus("db", dbStatus),
					}
					
					for name, svc := range desc.Services {
						if name != "web" && name != "db" && (svc.Status == "running" || svc.Status == "starting" || svc.Status == "unhealthy") {
							containerStatuses = append(containerStatuses, formatServiceStatus(name, svc.Status))
						}
					}
					
					if desc.MutagenEnabled {
						containerStatuses = append(containerStatuses, "mutagen: "+statusOKStyle.Render("enabled"))
					}
					
					details += fmt.Sprintf("%s  %s\n", boldLabel.Render("Containers:"), strings.Join(containerStatuses, "   "))
				} else {
					details += fmt.Sprintf("%s %-12s %s %-14s %s %s\n", 
						boldLabel.Render("Name:"), p.Name,
						boldLabel.Render("Type:"), p.Type,
						boldLabel.Render("Path:"), p.AppRoot,
					)
				}
				
				details += fmt.Sprintf("%s  %s\n",
					boldLabel.Render("URL:"), p.PrimaryURL,
				)
				
				lastError := m.lastErrors[p.Name]
				logs := m.accumulatedLogs[p.Name]
				
				if lastError != "" {
					errLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render("Error:")
					formattedErr := formatError(lastError, 3)
					errLines := strings.Split(formattedErr, "\n")
					details += fmt.Sprintf("%s %s\n", errLabel, errLines[0])
					for _, line := range errLines[1:] {
						details += fmt.Sprintf("         %s\n", line)
					}
				} else if len(logs) > 0 {
					progressLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true).Render("Progress:")
					rule := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("-", 40))
					details += fmt.Sprintf("%s\n%s\n", progressLabel, rule)
					start := len(logs) - 3
					if start < 0 {
						start = 0
					}
					for idx := start; idx < len(logs); idx++ {
						details += fmt.Sprintf("%s\n", logs[idx])
					}
				} else if p.Status == "unhealthy" {
					tipLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true).Render("Unhealthy:")
					details += fmt.Sprintf("%s Project is unhealthy. Try toggling it to stop/restart, or press 'p' to poweroff DDEV globally.\n", tipLabel)
				} else if p.Status == "running" || p.Status == "OK" {
					okLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true).Render("Running:")
					details += fmt.Sprintf("%s at %s\n", okLabel, p.PrimaryURL)
				} else {
					stoppedLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true).Render("Stopped:")
					details += fmt.Sprintf("%s Project is stopped. Toggle (enter/space) to start it.\n", stoppedLabel)
				}
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
		config:             cfg,
		spinner:            sp,
		isRefreshing:       true,
		refreshMsg:         "Loading DDEV projects...",
		lastErrors:         make(map[string]string),
		descriptions:       make(map[string]ddev.DescribeRaw),
		latestLogs:         make(map[string]string),
		accumulatedLogs:    make(map[string][]string),
		isPoweringOff:      false,
		poweroffLogs:       nil,
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
