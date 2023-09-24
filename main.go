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
  if m.cache[m.notes[m.cursor].name] != "" {
    return m.cache[m.notes[m.cursor].name]
  }

  // Read and render the file
  content, err := os.ReadFile(notesFolder + m.notes[m.cursor].name)
  if err != nil {
    log.Fatal(err)
  }
  out, err := m.renderer.Render(string(content))
  if err != nil {
    log.Fatal(err)
  }

  // Cache the file
  m.cache[m.notes[m.cursor].name] = out

  return out
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { 
  var cmd tea.Cmd

  switch msg := msg.(type) {
  case tea.KeyMsg:
    switch msg.String() {
    case "esc":
      if m.table.Focused() {
        m.table.Blur()
      } else {
        m.table.Focus()
      }
    case "q", "ctrl+c":
      return m, tea.Quit
    case "enter":
      m.openNote(m.table.SelectedRow()[0])
    }
  }

  // switch msg := msg.(type) {
  // case tea.KeyMsg:
  //   switch msg.String() {
  //   case "ctrl+c", "q", "esc":	
  //     return m, tea.Quit
  //   case "up", "k":
  //     if m.cursor > 0 {
  //       m.cursor--
  //       m.noteListView.LineUp(1)
  //     }
  //   case "down", "j":
  //     if m.cursor < len(m.notes)-1 {
  //       m.cursor++
  //       m.noteListView.LineDown(1)
  //     }
  //   case "n":
  //     m.openNote("untitled.md")
  //     return m, tea.Quit
      
  //   case "enter":
  //     if(len(m.notes) == 0) {
  //       return m, nil
  //     }
  //     m.openNote(m.notes[m.cursor].name)
  //     return m, nil
  //   }

  //   if msg.String()== "k" ||  msg.String() == "j" || msg.String() == "up" || msg.String() == "down" {
  // 		m.noteView.SetContent(m.renderNote())
  // 		m.noteListView.SetContent(m.renderNotes())
  // 	}

  // 	return m, nil
  // case tea.WindowSizeMsg:
  //   // Initialize viewports
  //   if !m.ready {
  //     m.noteView = viewport.New(msg.Width, msg.Height * 4/5)
  //     m.noteView.SetContent(m.renderNote())
  //     m.noteListView = viewport.New(msg.Width, msg.Height * 1/5)
  //     m.noteListView.SetContent(m.renderNotes())

  //     m.ready = true
  //   // Adjust viewports on terminal resize
  //   } else {
  //     m.noteView.Width = msg.Width
  //     m.noteView.Height = msg.Height / 4/5
  //     m.noteListView.Width = msg.Width
  //     m.noteListView.Height =  msg.Height * 1/5
  //   }
  // }

  m.table, cmd = m.table.Update(msg)
  return m, cmd
}

func (m model) View() string {
  return fmt.Sprintf("%s", baseStyle.Render(m.table.View()))
  // return fmt.Sprintf("%s%s", m.noteView.View(), m.noteListView.View())
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
  )

  if _, err := p.Run(); err != nil {
    fmt.Printf("Alas, there's been an error: %v", err)
    os.Exit(1)
  }
}