package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().Foreground(getTextStyle().GetForeground())
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4).Foreground(getTextStyle().GetForeground())
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(getTextStyle().GetForeground())
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

const defaultListWidth = 60

type versionItemDelegate struct{}

func (d versionItemDelegate) Height() int                               { return 1 }
func (d versionItemDelegate) Spacing() int                              { return 0 }
func (d versionItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d versionItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(version)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s - %s", index+1, i.title, i.description)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprintf(w, "%v", fn(str))
}

func createList(items []list.Item) list.Model {
	l := list.NewModel(items, versionItemDelegate{}, defaultListWidth, listHeight)
	l.Title = "What semver increment do you want to do?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	return l
}
