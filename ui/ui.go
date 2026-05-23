package ui

import (
	"fmt"
	"os/user"

	"ddev-clim/config"
	"ddev-clim/ddev"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)
	statusRunning = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	statusStopped = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	statusWorking = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	instructionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
)

type item struct {
	project    ddev.Project
	processing bool
	spinner    string // Current spinner frame
}

func (i item) Title() string { return i.project.Name }
func (i item) Description() string {
	if i.processing {
		return statusWorking.Render(i.spinner+" processing...") + " - " + i.project.AppRoot
	}
	status := i.project.Status
	if status == "running" || status == "OK" {
		return statusRunning.Render("● "+status) + " - " + i.project.AppRoot
	}
	return statusStopped.Render("○ "+status) + " - " + i.project.AppRoot
}
func (i item) FilterValue() string { return i.project.Name }

type statusMsg struct {
	index int
	err   error
}

type refreshMsg struct {
	projects []ddev.Project
}

type model struct {
	list           list.Model
	config         *config.Config
	spinner        spinner.Model
	filepicker     filepicker.Model
	showFilePicker bool
	err            error
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showFilePicker {
		var cmd tea.Cmd
		m.filepicker, cmd = m.filepicker.Update(msg)

		// Check if a directory was selected
		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			m.config.ScanPath = path
			m.showFilePicker = false
			_ = config.SaveConfig(m.config)
			return m, tea.Batch(m.refreshCmd(), m.spinner.Tick)
		}

		// Handle quit or back from picker
		if msg, ok := msg.(tea.KeyMsg); ok {
			if msg.String() == "esc" || msg.String() == "q" {
				m.showFilePicker = false
				return m, nil
			}
		}

		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "p" {
			m.showFilePicker = true
			return m, m.filepicker.Init()
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
				m.list.SetItem(idx, i)
			}
		}
		return m, cmd

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.filepicker.Height = msg.Height - v - 6 // Set height for filepicker
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		var projects []ddev.Project
		var err error
		if m.config.ScanPath != "" {
			projects, err = ddev.ScanForProjects(m.config.ScanPath)
		} else {
			projects, err = ddev.GetProjects()
		}
		if err != nil {
			return nil
		}
		return refreshMsg{projects: projects}
	}
}

func (m model) View() string {
	if m.showFilePicker {
		return docStyle.Render("Select a folder to scan for DDEV projects:\n\n" + m.filepicker.View() + "\n\n(esc to cancel)")
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

	var projects []ddev.Project
	if cfg.ScanPath != "" {
		projects, err = ddev.ScanForProjects(cfg.ScanPath)
	} else {
		projects, err = ddev.GetProjects()
	}
	if err != nil {
		return err
	}

	items := []list.Item{}
	for _, p := range projects {
		items = append(items, item{project: p})
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	fp := filepicker.New()
	fp.DirAllowed = true
	fp.FileAllowed = false
	fp.ShowHidden = false
	fp.Height = 10 // Default height
	usr, _ := user.Current()
	fp.CurrentDirectory = usr.HomeDir

	m := model{
		config:     cfg,
		spinner:    sp,
		filepicker: fp,
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Styles.Title = lipgloss.NewStyle() // Unset default title styling
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "toggle"),
			),
			key.NewBinding(
				key.WithKeys("p"),
				key.WithHelp("p", "path"),
			),
		}
	}
	// We handle title styling manually to avoid styling instructions
	
	instr := "\n" + instructionStyle.Render("Navigate the instances and toggle on and off")
	l.Title = titleStyle.Render("DDEV CLInstance Manager") + instr

	
	m.list = l

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
