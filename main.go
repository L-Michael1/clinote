package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
  notesFolder string
  editor []string

  baseStyle = lipgloss.NewStyle().
    BorderStyle(lipgloss.NormalBorder()).
    BorderForeground(lipgloss.Color("240"))

  titleStyle = func() lipgloss.Style {
    b := lipgloss.RoundedBorder()
    b.Right = "├"
    return lipgloss.NewStyle().Foreground(lipgloss.Color("225")).BorderStyle(b).Padding(0, 1)
  }()

  infoStyle = func() lipgloss.Style {
    b := lipgloss.RoundedBorder()
    b.Left = "┤"
    return titleStyle.Copy().BorderStyle(b)
  }()

  keys = keyMap{
    Up: key.NewBinding(
      key.WithKeys("up", "k"),
      key.WithHelp("↑/k", "move up"),
    ),
    Down: key.NewBinding(
      key.WithKeys("down", "j"),
      key.WithHelp("↓/j", "move down"),
    ),
    Help: key.NewBinding(
      key.WithKeys("?", "/"),
      key.WithHelp("?", "toggle help"),
    ),
    Quit: key.NewBinding(
      key.WithKeys("q", "ctrl+c"),
      key.WithHelp("q", "quit"),
    ),
    Back: key.NewBinding(
      key.WithKeys("b", "backspace"),
      key.WithHelp("b/bspace/esc", "back"),
    ),
    View: key.NewBinding(
      key.WithKeys("enter"),
      key.WithHelp("enter", "view note"),
    ),
    New: key.NewBinding(
      key.WithKeys("n"),
      key.WithHelp("n", "new note"),
    ),
    Edit: key.NewBinding(
      key.WithKeys("e"),
      key.WithHelp("e", "edit note"),
    ),
  }
)

type keyMap struct {
  Up    key.Binding
  Down  key.Binding
  Help  key.Binding
  Quit  key.Binding
  Back  key.Binding
  View  key.Binding
  New   key.Binding
  Edit  key.Binding
}

type note struct {
  name            string
  timeModified    string
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
  keys            keyMap
  help            help.Model
}

func (m model) Init() tea.Cmd { 
  return nil 
}

func (k keyMap) ShortHelp() []key.Binding {
  return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
  return [][]key.Binding{
    {k.Up, k.Down, k.Back},
    {k.View, k.New, k.Edit},      
    {k.Help, k.Quit},               
  }
}

func (m model) openNote(fileName string) {
  editor = strings.Split(os.Getenv("EDITOR"), " ")

  editor = append(editor, notesFolder + fileName)
  args := editor[1:]

  cmd := exec.Command(
    editor[0],
    args...)

  cmd.Stdin = os.Stdin
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr

  err := cmd.Start()
  checkErr(err)

  err = cmd.Wait()
  checkErr(err)
}

func (m model) renderNote() string {
  if(len(m.notes) == 0) {
    out, err := m.renderer.Render("# No notes found in " + notesFolder + ".\n\nPress `n` to create a new note.\n")
    checkErr(err)
    return out
  }

  // Read and render the file
  content, err := os.ReadFile(notesFolder + m.note)
  checkErr(err)

  out, err := m.renderer.Render(string(content))
  checkErr(err)

  return out
}

// Main update loop
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { 
  var cmd tea.Cmd
  
  switch msg := msg.(type) {
  case tea.KeyMsg:
    
    // Help menu toggle
    switch {
    case key.Matches(msg, m.keys.Help):
      m.help.ShowAll = !m.help.ShowAll
    }

    switch msg.String() {

    // Exit program
    case "ctrl+c", "q":
      return m, tea.Quit
    
    // Create new note in editor
    case "n":
      m.openNote("new_note.md")
      resetTerminal()
      return m, cmd

    // Back to table view
    case "b", "esc", "backspace":
      m.chosen = false
      m.table.Focus()
      m.table, cmd = m.table.Update(msg)
      return m, cmd
    }

  case tea.WindowSizeMsg:
    // Note view
    headerHeight := lipgloss.Height(m.headerView())
    footerHeight := lipgloss.Height(m.footerView())
    verticalMarginHeight := headerHeight + footerHeight
    
    // Help view
    m.help.Width = msg.Width

    // Wait asynchronously for the window size to be available
    if !m.ready {
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

  // Call update loop on table
  if !m.chosen {
    m.table.Focus() 
    return updateTableView(msg, m)
  } 

  // Call update loop on note
  return updateNoteView(msg, m)
}

func (m model) View() string {
  var s string

  if len(m.notes) == 0 {
    s = noNotesView(m)
  } else {
    if !m.chosen {
      s = tableView(m)
    } else {
      s = noteView(m)
    }
  }

  return s
}

// Sub-update functions

// Update loop for the choices (table) view
func updateTableView(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
  var cmd tea.Cmd

  switch msg := msg.(type) {
  case tea.KeyMsg:
    switch msg.String() {
    case "enter":
      if len(m.notes) != 0 {

        // Go to note view
        m.chosen = true
        m.note = m.table.SelectedRow()[0]
        m.table.Blur()
        m.noteView.SetContent(m.renderNote())
        return m, nil
      }

    case "e":
      if len(m.notes) != 0 {
        // Open the selected note for editing
        m.openNote(m.table.SelectedRow()[0])
        resetTerminal()
      }
    }
  }
  
  m.table, cmd = m.table.Update(msg)

  return m, cmd
}

// Update loop for the note view after a note has been chosen
func updateNoteView(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
  var (
    cmd  tea.Cmd
    cmds []tea.Cmd
  )

  // Handle keyboard and mouse events in the viewport
  m.noteView, cmd = m.noteView.Update(msg)
  cmds = append(cmds, cmd)

  switch msg := msg.(type) {
  case tea.KeyMsg:
    switch msg.String() {
    case "e":
      // Open the current note in the editor for editing
      m.openNote(m.note)

      // Return to the table view
      m.chosen = false
      m.table.Focus()
      m.table, cmd = m.table.Update(msg)
      return m, cmd
    }
  }

  return m, tea.Batch(cmds...)
}

// Sub-view functions 

func noNotesView(m model) string {
  return fmt.Sprintf("%s", baseStyle.Render(lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Render("No notes found in " + notesFolder + ". Press `n` to create a new note!")))
}

func tableView(m model) string {
  helpView := m.help.View(m.keys)
  return fmt.Sprintf("%s\n%s", baseStyle.Render(m.table.View()), helpView)
}

func noteView(m model) string {
  if !m.ready {
    return "\n  Initializing..."
  }
  return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.noteView.View(), m.footerView())
}

func (m model) headerView() string {
  if m.table.SelectedRow() != nil {
    title := titleStyle.Render(m.table.SelectedRow()[0])
    line := strings.Repeat("─", max(0, m.noteView.Width-lipgloss.Width(title)))
    return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
  }
  
  return ""
}

func (m model) footerView() string {
  info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.noteView.ScrollPercent()*100))
  line := strings.Repeat("─", max(0, m.noteView.Width-lipgloss.Width(info)))
  return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

// Utils

func getNotes() []note {
  var notes []note
  const timeFormat = "2006-01-02 15:04"
  entries, err := os.ReadDir(notesFolder)
  checkErr(err)

  // Add notes to our model
  for _, entry := range entries {

    // Check if entry is a file
    if !entry.IsDir() {
      fileInfo, err := os.Stat(notesFolder + entry.Name())
      checkErr(err)

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

func (m model) updateTable() (table.Model, tea.Cmd) {
  rows := convertNotesToRows(m.notes)
  m.table.SetRows(rows)
  return m.table, nil
}

func resetTerminal() {
  cmd := exec.Command("clinote")
  cmd.Stdin = os.Stdin
  cmd.Stdout = os.Stdout
  cmd.Run()
}

func checkErr(err error) {
  if err != nil {
    log.Fatal(err)
  }
}

func main() {
  
  // If no notes folder is specified, default to ~/notes/
  if os.Getenv("NOTES_FOLDER") == "" {
    notesFolder = os.Getenv("HOME") + "/notes/"
  } else {
    notesFolder = os.Getenv("NOTES_FOLDER") + "/"
  }

  // If no editor is specified, default to vim
  if os.Getenv("EDITOR") == "" {
    os.Setenv("EDITOR", "vim")
  } else {
    editor = strings.Split(os.Getenv("EDITOR"), " ")
  }

  // Table configuration
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
    BorderStyle(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("240")).
    BorderBottom(true)
  s.Selected = s.Selected.
    Foreground(lipgloss.Color("213")).
    Background(lipgloss.Color("25")).
    Bold(true)
  t.SetStyles(s)

  terminalWidth, _, err := term.GetSize(0)
  checkErr(err)

  renderer, err := glamour.NewTermRenderer(
      glamour.WithAutoStyle(),
      glamour.WithWordWrap(terminalWidth),
  )
  checkErr(err)
  
  p := tea.NewProgram(
    model {
      note:       string(""),
      notes:      notes,
      table:      t,
      cache:      make(map[string]string, len(notes)),
      renderer:   *renderer,
      keys:       keys,
      help:       help.New(),
    },
    tea.WithAltScreen(),
    tea.WithMouseCellMotion(),
  )

  if _, err := p.Run(); err != nil {
    fmt.Printf("Alas, there's been an error: %v", err)
    os.Exit(1)
  }
}