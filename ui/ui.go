package ui

import (
	"fmt"
	"io"
	"strings"

	"ddev-clim/config"
	"ddev-clim/ddev"

	"github.com/charmbracelet/bubbles/key"
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
	index int
	err   error
}

type refreshMsg struct {
	projects []ddev.Project
}

type model struct {
	list         list.Model
	config       *config.Config
	spinner      spinner.Model
	isRefreshing bool
	refreshMsg   string
	err          error
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

					if i.project.Status == "running" || i.project.Status == "OK" || i.project.Status == "unhealthy" {
						m.config.RemoveProject(i.project.Name)
					} else {
						m.config.AddProject(i.project.Name)
					}
					_ = config.SaveConfig(m.config)

					return m, func() tea.Msg {
						var err error
						if i.project.Status == "running" || i.project.Status == "OK" || i.project.Status == "unhealthy" {
							err = ddev.StopProject(i.project.AppRoot)
						} else {
							err = ddev.StartProject(i.project.AppRoot)
						}
						return statusMsg{index: idx, err: err}
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
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.err = nil
		}
		m.refreshMsg = "Loading DDEV projects..."
		return m, m.refreshCmd()

	case refreshMsg:
		m.isRefreshing = false
		items := []list.Item{}
		for _, p := range msg.projects {
			items = append(items, item{project: p})
		}
		m.list.SetItems(items)
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
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		projects, err := ddev.GetProjects()
		if err != nil {
			return nil
		}
		return refreshMsg{projects: projects}
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
	
	// Check for unhealthy instances
	hasUnhealthy := false
	for _, listItem := range m.list.Items() {
		if i, ok := listItem.(item); ok && i.project.Status == "unhealthy" {
			hasUnhealthy = true
			break
		}
	}
	
	if hasUnhealthy {
		warningMsg := "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true).Render(
			"⚠️ Unhealthy instance(s) detected! Try toggling to stop/restart them, or press 'p' to poweroff DDEV.",
		)
		s += warningMsg
	}

	if m.err != nil {
		s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("Error: %v", m.err))
	}
	
	return s
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
	}

	d := itemDelegate{spinner: sp}

	l := list.New([]list.Item{}, d, 0, 0)
	l.Styles.Title = lipgloss.NewStyle() // Unset default title styling
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("enter", " "),
				key.WithHelp("enter/space", "toggle"),
			),
			key.NewBinding(
				key.WithKeys("p"),
				key.WithHelp("p", "poweroff"),
			),
		}
	}
	
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
