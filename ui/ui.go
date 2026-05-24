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
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)
	statusRunning    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	statusStopped    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	statusWorking    = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	instructionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	pathStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	headerStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Bold(true).Underline(true)
	fuchsia          = lipgloss.Color("205")
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
	
	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	if index == m.Index() {
		nameStyle = nameStyle.Foreground(fuchsia)
	}
	
	nameStr := runewidth.FillRight(name, nameWidth)
	str += nameStyle.Render(nameStr)

	// Status
	status := i.project.Status
	var statusStr string
	if i.processing {
		statusStr = statusWorking.Render(i.spinner + " starting...")
	} else if status == "running" || status == "OK" {
		statusStr = statusRunning.Render(status)
	} else {
		statusStr = statusStopped.Render(status)
	}
	
	// Manual padding for status
	rawStatus := status
	if i.processing {
		rawStatus = "starting..."
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

				return m, func() tea.Msg {
					var err error
					if i.project.Status == "running" || i.project.Status == "OK" {
						err = ddev.StopProject(i.project.AppRoot)
					} else {
						err = ddev.StartProject(i.project.AppRoot)
					}
					return statusMsg{index: idx, err: err}
				}
			}
		}
	case statusMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		return m, m.refreshCmd()

	case refreshMsg:
		m.isRefreshing = false
		items := []list.Item{}
		running := []string{}
		for _, p := range msg.projects {
			items = append(items, item{project: p})
			if p.Status == "running" || p.Status == "OK" {
				running = append(running, p.Name)
			}
		}
		m.list.SetItems(items)
		m.config.RunningProjects = running
		_ = config.SaveConfig(m.config)
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
		return docStyle.Render(fmt.Sprintf("\n\n  %s Loading DDEV projects...", m.spinner.View()))
	}

	s := docStyle.Render(m.list.View())
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
	}

	d := itemDelegate{spinner: sp}

	l := list.New([]list.Item{}, d, 0, 0)
	l.Styles.Title = lipgloss.NewStyle() // Unset default title styling
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "toggle"),
			),
		}
	}
	
	instr := "\n" + instructionStyle.Render("Navigate the instances and toggle on and off")
	
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
