package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var notesFolder = os.Getenv("NOTES_FOLDER") + "/"

var baseStyle = lipgloss.NewStyle().
  BorderStyle(lipgloss.NormalBorder()).
  BorderForeground(lipgloss.Color("240"))

var (
  titleStyle = func() lipgloss.Style {
    b := lipgloss.RoundedBorder()
    b.Right = "├"
    return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
  }()

  infoStyle = func() lipgloss.Style {
    b := lipgloss.RoundedBorder()
    b.Left = "┤"
    return titleStyle.Copy().BorderStyle(b)
  }()
)

type note struct {
  name          string
  timeModified  string
}
type model struct {
  note 			      string
  notes 		      []note
  table 			    table.Model
  cursor		      int
  ready           bool
  renderer 	      glamour.TermRenderer
  noteView        viewport.Model
  cache           map[string]string
  chosen          bool
}

func (m model) Init() tea.Cmd { 
  return nil 
}

func (m model) openNote(fileName string) {
  editor := strings.Split(os.Getenv("EDITOR"), " ")

  editor = append(editor, notesFolder + fileName)
  args := editor[1:]

  cmd := exec.Command(
    editor[0],
    args...)

  cmd.Stdin = os.Stdin
  cmd.Stdout = os.Stdout

  err := cmd.Run()
  if err != nil {
    log.Fatal(err)
  }
}

func (m model) renderNote() string {
  // If there are no notes, display a message
  if(len(m.notes) == 0) {
    out, err := m.renderer.Render("# No notes found in " + notesFolder + ".\n\nPress `n` to create a new note.\n")
    if err != nil {
      log.Fatal(err)
    }
    return out
  }

  // Check if the file is cached
  // if m.cache[m.note] != "" {
  //   return m.cache[m.note]
  // }

  // Read and render the file
  content, err := os.ReadFile(notesFolder + m.note)
  if err != nil {
    log.Fatal(err)
  }
  out, err := m.renderer.Render(string(content))
  if err != nil {
    log.Fatal(err)
  }

  // Cache the file
  // m.cache[m.note] = out

  return out
}

// Main update function
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { 
  var cmd tea.Cmd
	
  switch msg := msg.(type) {
  case tea.KeyMsg:
    switch msg.String() {
    case "q", "ctrl+c":
      return m, tea.Quit
    case "b":
      m.chosen = false
      m.table.Focus()
      m.table, cmd = m.table.Update(msg)
      return m, cmd
    }

  case tea.WindowSizeMsg:
    headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

    if !m.ready {
      // Wait asynchronously for the window size to be available
			m.noteView = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.noteView.YPosition = headerHeight
			m.ready = true

			// Render the viewport one line below the header.
			m.noteView.YPosition = headerHeight + 1
		} else {
			m.noteView.Width = msg.Width
			m.noteView.Height = msg.Height - verticalMarginHeight
		}
	}

  // Update table if note not chosen
  if !m.chosen {
    return updateChoices(msg, m)
  } 

  // Update note if chosen
  return updateChosen(msg, m)
}

func (m model) View() string {
  var s string

  if !m.chosen {
    s = tableView(m)
  } else {
    s = noteView(m)
  }

  return s
}

// Sub-update functions

// Update loop for the choices (table) view
func updateChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
  var cmd tea.Cmd

  switch msg := msg.(type) {
  case tea.KeyMsg:
    switch msg.String() {
    case "enter":
      m.chosen = true
      m.note = m.table.SelectedRow()[0]
      m.noteView.SetContent(m.renderNote())
      return m, tea.Batch(tea.Printf("Let's go to %s!", m.table.SelectedRow()[0]))
    }
  }
  
  m.table, cmd = m.table.Update(msg)

  return m, cmd
}

// Update loop for the second view after a choice has been made
func updateChosen(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
  var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
    switch msg.String() {
    case "e":
      fmt.Println("edit")
      m.openNote(m.table.SelectedRow()[0])
      return m, nil
    }
  }

  // Handle keyboard and mouse events in the viewport
	m.noteView, cmd = m.noteView.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// Sub-view functions 

func tableView(m model) string {
  return fmt.Sprintf("%s", baseStyle.Render(m.table.View()))
}

func noteView(m model) string {
  if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.noteView.View(), m.footerView())
}

func (m model) headerView() string {
	title := titleStyle.Render(m.table.SelectedRow()[0])
	line := strings.Repeat("─", max(0, m.noteView.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.noteView.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.noteView.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func getInitialContent(notes []fs.DirEntry) string {
  // Load the first note if there is one
  var content string
  if len(notes) > 0 {
    contentBytes, err := os.ReadFile(notesFolder + notes[0].Name())
    if err != nil {
      log.Fatal(err)
    }

    content = string(contentBytes)
  }

  return content
}

func getNotes() []note {
  var notes []note
  timeFormat := "2006-01-02 15:04"
  entries, err := os.ReadDir(notesFolder)
  if err != nil {
    log.Fatal(err)
  }

  // Add notes to our model
  for _, entry := range entries {

    // Check if entry is a file
    if !entry.IsDir() {
      fileInfo, err := os.Stat(notesFolder + entry.Name())
      if err != nil {
        log.Fatal(err)
      }

      fileName := fileInfo.Name()
      timeModified := fileInfo.ModTime().Format(timeFormat)
      note := note {
        name: fileName,
        timeModified: timeModified,
      }

      notes = append(notes, note)
    }
  }

  return notes
}

func convertNotesToRows(notes []note) []table.Row {
  var rows []table.Row

  for _, row := range notes {
    rows = append(rows, table.Row {
      row.name,
      row.timeModified,
    })
  }

  return rows
}

func main() {
  columns := []table.Column{
    {Title: "Note", Width: 25},
    {Title: "Date Modified", Width: 16},
  }

  notes := getNotes()
  rows := convertNotesToRows(notes)
 
  t := table.New(
    table.WithColumns(columns),
    table.WithRows(rows),
    table.WithFocused(true),
    table.WithHeight(7),
  )

  s := table.DefaultStyles()
  s.Header = s.Header.
    BorderStyle(lipgloss.NormalBorder()).
    BorderForeground(lipgloss.Color("240")).
    BorderBottom(true).
    Bold(false)
  s.Selected = s.Selected.
    Foreground(lipgloss.Color("229")).
    Background(lipgloss.Color("57")).
    Bold(false)
  t.SetStyles(s)

  terminalWidth, _, err := term.GetSize(0)

  renderer, err := glamour.NewTermRenderer(
      glamour.WithAutoStyle(),
      glamour.WithWordWrap(terminalWidth-3))
  if err != nil {
    log.Fatal(err)
  }
  
  p := tea.NewProgram(
    model {
      note:  string(""),
      notes: notes,
      table: t,
      cache:  make(map[string]string, len(notes)),
      renderer: *renderer,
    },
    tea.WithAltScreen(),
    tea.WithMouseCellMotion(),
  )

  if _, err := p.Run(); err != nil {
    fmt.Printf("Alas, there's been an error: %v", err)
    os.Exit(1)
  }
}