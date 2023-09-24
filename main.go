package main

import (
	"fmt"
	"io/fs"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"golang.org/x/term"
)

type model struct {
	note 			string
	notes 		[]fs.DirEntry
	renderer 	glamour.TermRenderer
}

func (m model) Init() tea.Cmd { 
	return nil 
}

func (m model) Update() tea.Cmd { 
	return nil
}

func (m model) View() string { 
	return "" 
}

func main() {
	
	// Test Markdown
	test := `# Hello World

This is a simple example of Markdown rendering with Glamour!
Check out the [other examples](https://github.com/charmbracelet/glamour/tree/master/examples) too.

Bye!
`

	terminalWidth, _, err := term.GetSize(0)

	renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(terminalWidth-5))
	if err != nil {
		log.Fatal(err)
	}
	
	out, _ := renderer.Render(test)

	fmt.Print(out)
}