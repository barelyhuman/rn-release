package main

// A simple program demonstrating the spinner component from the Bubbles
// component library.

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	spinner             spinner.Model
	quitting            bool
	err                 error
	processName         string
	selectedSemver      string
	showSemver          bool
	semverList          list.Model
	createScripts       bool
	textInput           textinput.Model
	showPodInput        bool
	showGradleInput     bool
	podFileLocation     string
	buildGradleLocation string
}

type AppState int64

type AppStateMsg AppState

type item string

func (i item) Title() string       { return string(i) }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return string(i) }

var listStyle = lipgloss.NewStyle().Margin(1, 2)

const (
	Init AppState = iota
	Initialized
	CheckFiles
	FilesChecked
	FilesNotFound
	SearchPlatformFiles
	PlatformFilesFound
	CreateScripts
	CreatedScripts
	ShowSemverSelect
	SemVerIncreased
	SyncingVersion
	GetPodFile
	GetGradleFile
	Done
)

const CONFIG_FOLDER = ".rnrelease"
const SCRIPT_FILE_NAME = "sync_version.sh"

var SEMVER_COMMANDS = []string{
	"patch",
	"minor",
	"major",
	"prepatch",
	"preminor",
	"premajor",
	"prerelease",
}

func pause() {
	time.Sleep(2 * time.Second)
}

func startInit() tea.Msg {
	bail(os.MkdirAll(CONFIG_FOLDER, os.ModePerm))
	pause()
	return AppStateMsg(Initialized)
}

func createInput() textinput.Model {
	ti := textinput.NewModel()
	ti.Placeholder = "..."
	ti.TextStyle.Foreground(getTextStyle().GetForeground())
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	return ti
}

func createSpinner() spinner.Model {
	s := spinner.NewModel()
	s.Spinner = spinner.Line
	s.Style = lipgloss.NewStyle().Foreground(getTextStyle().GetForeground())
	return s
}

func createList(items []list.Item) list.Model {
	l := list.NewModel(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "What semver increment do you want to do?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = getTextStyle()
	l.Styles.PaginationStyle = getTextStyle()
	l.Styles.HelpStyle = getTextStyle()
	return l
}

func initialModel() model {

	ti := createInput()
	s := createSpinner()

	items := []list.Item{}

	for _, x := range SEMVER_COMMANDS {
		items = append(items, item(x))
	}

	l := createList(items)

	return model{spinner: s, semverList: l, textInput: ti}
}

func appInit() tea.Msg {
	return AppStateMsg(Init)
}

func startFileCheck() tea.Msg {
	return AppStateMsg(CheckFiles)
}

func checkExistingFiles() tea.Msg {
	files, err := os.ReadDir(CONFIG_FOLDER)
	bail(err)
	createScript := true
	for _, file := range files {
		if file.Name() == SCRIPT_FILE_NAME {
			createScript = false
			break
		}
	}
	if createScript {
		return AppStateMsg(FilesNotFound)
	}
	return AppStateMsg(FilesChecked)
}

func getPodFileLocation() tea.Msg {
	return AppStateMsg(GetPodFile)
}
func getGradleFileLocation() tea.Msg {
	return AppStateMsg(GetGradleFile)
}

func startCreatingScript() tea.Msg {
	return AppStateMsg(CreateScripts)
}

func startVersionCheck() tea.Msg {
	pause()
	return AppStateMsg(ShowSemverSelect)
}

func createScripts() tea.Msg {
	// write to disk
	pause()
	return AppStateMsg(CreatedScripts)
}

func startSemverIncrement() tea.Msg {
	// increase semver using `npm`
	return AppStateMsg(SemVerIncreased)
}

func syncWithPlatform() tea.Msg {
	// run the generated script
	pause()
	return AppStateMsg(Done)
}

func (m model) Init() tea.Cmd {
	return tea.Batch(spinner.Tick, appInit)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		top, right, bottom, left := listStyle.GetMargin()
		m.semverList.SetSize(msg.Width-left-right, msg.Height-top-bottom)
		var cmd tea.Cmd
		m.semverList, cmd = m.semverList.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":

			if m.createScripts && m.showPodInput {
				m.podFileLocation = m.textInput.Value()
				m.textInput.SetValue("")
				m.showPodInput = false
				return m, getGradleFileLocation
			}

			if m.createScripts && m.showGradleInput {
				m.buildGradleLocation = m.textInput.Value()
				m.textInput.SetValue("")
				m.showGradleInput = false
				return m, startCreatingScript
			}

			if m.showSemver {
				i, ok := m.semverList.SelectedItem().(item)
				if ok {
					m.selectedSemver = string(i)
					return m, startSemverIncrement
				}
			}

		default:
			var cmds []tea.Cmd
			var c tea.Cmd

			m.semverList, c = m.semverList.Update(msg)
			cmds = append(cmds, c)

			m.textInput, c = m.textInput.Update(msg)
			cmds = append(cmds, c)

			m.spinner, c = m.spinner.Update(msg)
			cmds = append(cmds, c)
			cmds = append(cmds, textinput.Blink)
			return m, tea.Batch(cmds...)
		}

	case AppStateMsg:
		switch msg {
		case AppStateMsg(Init):
			m.processName = "Initializing"
			return m, startInit
		case AppStateMsg(Initialized):
			m.processName = "Initialized"
			return m, startFileCheck
		case AppStateMsg(CheckFiles):
			m.processName = "Looking for existing files"
			return m, checkExistingFiles
		case AppStateMsg(FilesNotFound):
			m.createScripts = true
			return m, getPodFileLocation
		case AppStateMsg(GetPodFile):
			m.showPodInput = true
			return m, nil
		case AppStateMsg(GetGradleFile):
			m.showGradleInput = true
			return m, nil
		case AppStateMsg(CreateScripts):
			m.processName = "Creating Scripts"
			return m, createScripts
		case AppStateMsg(CreatedScripts):
			m.createScripts = false
			m.processName = "Looking for versioning data"
			return m, startVersionCheck
		case AppStateMsg(ShowSemverSelect):
			m.showSemver = true
			return m, nil
		case AppStateMsg(SemVerIncreased):
			m.showSemver = false
			m.processName = "Syncing versions with platform files"
			return m, syncWithPlatform
		case AppStateMsg(Done):
			return m, tea.Quit
		default:
			return m, nil
		}
	}
	var c tea.Cmd
	m.spinner, c = m.spinner.Update(msg)
	return m, c
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	if m.createScripts {
		if m.showPodInput {
			return fmt.Sprintf("\n   %s\n%s \n", getTextStyle().Render("Location of your Podfile"), m.textInput.View())

		}
		if m.showGradleInput {
			if m.buildGradleLocation == "" {
				return fmt.Sprintf("\n   %s\n%s \n", getTextStyle().Render("Location of your build.gradle"), m.textInput.View())
			}
		}
	}

	if m.showSemver {
		return fmt.Sprintf("\n   %s \n", m.semverList.View())
	}

	return fmt.Sprintf("\n   %s %s \n", m.spinner.View(), getTextStyle().Render(m.processName))

	// return fmt.Sprintf("\n   %s %s\n", getTextStyle().Render("✓"), getTextStyle().Render(activeProcess.name))

	// return fmt.Sprintf("\n   %s %s %s\n", m.spinner.View(), getTextStyle().Render(activeProcess.name), getTextStyle().Render(m.selectedSemver))
}

func GetUI() *tea.Program {
	p := tea.NewProgram(initialModel())
	return p
}
