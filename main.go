package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"golang.org/x/term"
)

var notesFolder = os.Getenv("NOTES_FOLDER") + "/"
type model struct {
  note 			      string
  notes 		      []fs.DirEntry
  cursor		      int
  ready           bool
  renderer 	      glamour.TermRenderer
  noteView        viewport.Model
  noteListView    viewport.Model
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
	if m.cache[m.notes[m.cursor].Name()] != "" {
		return m.cache[m.notes[m.cursor].Name()]
	}

	// Read and render the file
	content, err := os.ReadFile(notesFolder + m.notes[m.cursor].Name())
	if err != nil {
    log.Fatal(err)
  }
	out, err := m.renderer.Render(string(content))
	if err != nil {
    log.Fatal(err)
  }

	// Cache the file
	m.cache[m.notes[m.cursor].Name()] = out

	return out
}

func (m model) renderNotes() string {
	listStr := ""

	// Add cursor to the correct note
	for i, note := range m.notes {
		if note == nil {
			continue
		}

		if i == m.cursor {
			listStr += fmt.Sprintf("> %s\n", note.Name())
		} else {
			listStr += fmt.Sprintf("  %s\n", note.Name())
		}
	}

	return fmt.Sprintf(
		"\n\n  %s  \n\n%s",
		strings.Repeat(" ", m.noteListView.Width-4),
		listStr)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { 
  switch msg := msg.(type) {

  case tea.KeyMsg:
    switch msg.String() {
    case "ctrl+c", "q", "esc":	
      return m, tea.Quit
    case "up", "k":
      if m.cursor > 0 {
        m.cursor--
        m.noteListView.LineUp(1)
      }
    case "down", "j":
      if m.cursor < len(m.notes)-1 {
        m.cursor++
        m.noteListView.LineDown(1)
      }
    case "n":
      m.openNote("untitled.md")
      return m, tea.Quit
      
    case "enter":
      if(len(m.notes) == 0) {
        return m, nil
      }
      m.openNote(m.notes[m.cursor].Name())
      return m, nil
    }

    if msg.String()== "k" ||  msg.String() == "j" || msg.String() == "up" || msg.String() == "down" {
			m.noteView.SetContent(m.renderNote())
			m.noteListView.SetContent(m.renderNotes())
		}

		return m, nil
  case tea.WindowSizeMsg:
    // Initialize viewports
    if !m.ready {
      m.noteView = viewport.New(msg.Width, msg.Height * 4/5)
      m.noteView.SetContent(m.renderNote())
      m.noteListView = viewport.New(msg.Width, msg.Height * 1/5)
      m.noteListView.SetContent(m.renderNotes())

      m.ready = true
    // Adjust viewports on terminal resize
    } else {
      m.noteView.Width = msg.Width
      m.noteView.Height = msg.Height / 4/5
      m.noteListView.Width = msg.Width
      m.noteListView.Height =  msg.Height * 1/5
    }
  }
  
  return m, tea.Batch()
}

func (m model) View() string {
  return fmt.Sprintf("%s%s", m.noteView.View(), m.noteListView.View())
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

func getNotes() []fs.DirEntry {
  entries, err := os.ReadDir(notesFolder)
  var notes []fs.DirEntry
  if err != nil {
    log.Fatal(err)
  }

  // Add notes to our model
  for _, entry := range entries {

    // Check if entry is a file
    if !entry.IsDir() {
      notes = append(notes, entry)
    }
  }

  return notes
}

func main() {

  terminalWidth, _, err := term.GetSize(0)

  renderer, err := glamour.NewTermRenderer(
      glamour.WithAutoStyle(),
      glamour.WithWordWrap(terminalWidth-3))
  if err != nil {
    log.Fatal(err)
  }
  
  notes := getNotes()
	content := getInitialContent(notes)
  
  p := tea.NewProgram(
    model {
      note:  string(content),
      notes: notes,
      cache:  make(map[string]string, len(notes)),
      renderer: *renderer,
    },
  )

  if _, err := p.Run(); err != nil {
    fmt.Printf("Alas, there's been an error: %v", err)
    os.Exit(1)
  }
}