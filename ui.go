package main

// A simple program demonstrating the spinner component from the Bubbles
// component library.

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coreos/go-semver/semver"
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
	showPlistInput      bool
	showGradleInput     bool
	infoPlistLocation   string
	buildGradleLocation string
	version             string
}

type AppState int64

type AppStateMsg AppState

type version struct {
	title       string
	description string
}

type PackageJSON struct {
	Version string
}

func (i version) Title() string       { return i.title }
func (i version) Description() string { return i.description }
func (i version) FilterValue() string { return i.title }

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

var VERSIONING_FILES = []string{
	"package.json",
}

var SEMVER_COMMANDS = []string{
	"patch",
	"minor",
	"major",
	// TODO: re add them when you have better semver library that supports
	// bumping prerelease version, if not, then make one
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

func initialModel() model {

	ti := createInput()
	s := createSpinner()

	return model{spinner: s, textInput: ti}
}

func getVersionItems() []version {
	items := []version{}
	for _, x := range SEMVER_COMMANDS {
		items = append(items, version{
			title: x,
		})
	}
	return items
}

func versionToListItem(items []version) []list.Item {
	listItems := []list.Item{}
	for _, itemRef := range items {
		listItems = append(listItems, itemRef)
	}
	return listItems
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

func startVersionCheck(m model) (model, tea.Cmd) {
	var fileFound string
	for _, fileName := range VERSIONING_FILES {
		if _, err := os.Stat(fileName); err == nil {
			fileFound = fileName
		}
	}

	if fileFound == "" {
		verFileError := fmt.Errorf("no versioning file found, please add one of the following %v", VERSIONING_FILES)
		log.Fatal(verFileError)
		bail(verFileError)
	}

	switch fileFound {
	case "package.json":
		var buf PackageJSON
		fileData, err := os.ReadFile(fileFound)
		bail(err)
		bail(json.Unmarshal(fileData, &buf))
		m.version = buf.Version
	}

	pause()
	cmd := func() tea.Msg {
		return AppStateMsg(ShowSemverSelect)
	}
	return m, cmd
}

type ScriptInput struct {
	InfoPlistLocation   string
	BuildGradleLocation string
}

//go:embed templates/*.sh
var embedFS embed.FS

func createScriptFiles(m model) tea.Cmd {
	return func() tea.Msg {
		scriptOutput, err := os.Create(CONFIG_FOLDER + "/" + SCRIPT_FILE_NAME)
		bail(err)

		scriptOutput.Chmod(os.ModePerm)

		parsedTemplates, err := template.ParseFS(embedFS, "templates/*.sh")
		bail(err)
		defer scriptOutput.Close()

		err = parsedTemplates.ExecuteTemplate(scriptOutput, "syncVersionShell",
			ScriptInput{

				InfoPlistLocation:   m.infoPlistLocation,
				BuildGradleLocation: m.buildGradleLocation,
			})

		bail(err)

		pause()
		return AppStateMsg(CreatedScripts)
	}
}

func startSemverIncrement(m model) tea.Cmd {
	return func() tea.Msg {
		// increase semver using `npm`
		cmd := exec.Command("npm", "version", m.selectedSemver)
		// TODO: use the stdout to figure out the version
		// increased as intended
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		err := cmd.Run()

		if err != nil {
			log.Fatalf(stderr.String())
		}
		return AppStateMsg(SemVerIncreased)
	}
}

func syncWithPlatform() tea.Msg {
	// run the generated script
	cmd := exec.Command("./" + CONFIG_FOLDER + "/" + SCRIPT_FILE_NAME)
	_, err := cmd.Output()
	bail(err)
	pause()
	return AppStateMsg(Done)
}

func (m model) Init() tea.Cmd {
	return tea.Batch(spinner.Tick, appInit)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		if !m.showSemver {
			return m, nil
		}
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

			if m.createScripts && m.showPlistInput {
				m.infoPlistLocation = m.textInput.Value()
				m.textInput.SetValue("")
				m.showPlistInput = false
				return m, getGradleFileLocation
			}

			if m.createScripts && m.showGradleInput {
				m.buildGradleLocation = m.textInput.Value()
				m.textInput.SetValue("")
				m.showGradleInput = false
				return m, startCreatingScript
			}

			if m.showSemver {
				i, ok := m.semverList.SelectedItem().(version)
				if ok {
					m.selectedSemver = string(i.title)
					return m, startSemverIncrement(m)
				}
			}

		default:
			var cmds []tea.Cmd
			var c tea.Cmd

			if m.showSemver {
				m.semverList, c = m.semverList.Update(msg)
				cmds = append(cmds, c)
			}

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
			m.processName = "Check if script exists"
			return m, checkExistingFiles
		case AppStateMsg(FilesChecked):
			m.processName = "Looking for versioning data"
			m, cmd := startVersionCheck(m)
			return m, cmd
		case AppStateMsg(FilesNotFound):
			m.createScripts = true
			return m, getPodFileLocation
		case AppStateMsg(GetPodFile):
			m.showPlistInput = true
			return m, nil
		case AppStateMsg(GetGradleFile):
			m.showGradleInput = true
			return m, nil
		case AppStateMsg(CreateScripts):
			m.processName = "Creating Scripts"
			return m, createScriptFiles(m)
		case AppStateMsg(CreatedScripts):
			m.createScripts = false
			m.processName = "Looking for versioning data"
			m, cmd := startVersionCheck(m)
			return m, cmd
		case AppStateMsg(ShowSemverSelect):
			m.showSemver = true
			versionItems := getVersionItems()

			for i, versionItem := range versionItems {
				currentVersion := semver.New(m.version)
				switch versionItem.title {
				case "patch":
					patchV := semver.New(currentVersion.String())
					patchV.BumpPatch()
					newVer := (patchV.String())
					versionItems[i].description = newVer
				case "minor":
					minorV := semver.New(currentVersion.String())
					minorV.BumpMinor()
					versionItems[i].description = minorV.String()
				case "major":
					majorV := semver.New(currentVersion.String())
					majorV.BumpMajor()
					versionItems[i].description = majorV.String()
				default:
					versionItems[i].description = "Couldn't Predict"
				}
			}

			l := createList(versionToListItem(versionItems))
			m.semverList = l
			return m, nil
		case AppStateMsg(SemVerIncreased):
			m.showSemver = false
			m.processName = "Syncing version with platform files"
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
		if m.showPlistInput {
			return fmt.Sprintf("\n   %s\n%s \n", getTextStyle().Render("Location of your Info.plist"), m.textInput.View())

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
