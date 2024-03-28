package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/76creates/stickers"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// TODO(reno): enums are confusing, but I should strengthen types on the method
type model struct {
	flexbox       *stickers.FlexBox
	url           textinput.Model
	body          textarea.Model
	method        string
	response      viewport.Model
	r             string
	activeSection string
	editing       bool
	err           string
}

func initialModel() model {
	urlInput := textinput.New()
	urlInput.SetValue("http://localhost:9200")

	bodyInput := textarea.New()
	bodyInput.Placeholder = "Body"

	// TODO(reno): Look at the example code for viewport and get the height/width
	// that way. I'm just doing this lazily to try it.
	resp := viewport.New(0, 0)
	fb := stickers.NewFlexBox(0, 0)

	r1 := fb.NewRow().AddCells(
		[]*stickers.FlexBoxCell{
			stickers.NewFlexBoxCell(1, 1),
			stickers.NewFlexBoxCell(20, 1),
		},
	)
	r2 := fb.NewRow().AddCells(
		[]*stickers.FlexBoxCell{
			stickers.NewFlexBoxCell(1, 15),
			stickers.NewFlexBoxCell(1, 15),
		},
	)
	r3 := fb.NewRow().AddCells(
		[]*stickers.FlexBoxCell{
			stickers.NewFlexBoxCell(1, 1),
			stickers.NewFlexBoxCell(1, 1),
		},
	)

	fb.AddRows([]*stickers.FlexBoxRow{r1, r2, r3})

	return model{
		flexbox:       fb,
		url:           urlInput,
		body:          bodyInput,
		method:        "GET",
		response:      resp,
		r:             "",
		activeSection: "URL",
		editing:       false,
		err:           "",
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func makeRequest(m model) string {
	client := &http.Client{}
	// TODO(reno): body support, it's supposed to be of type io.Reader, not sure
	// how that works
	req, err := http.NewRequest(m.method, m.url.Value(), nil)
	check(err)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	check(err)
	body := make([]byte, 1000)
	resp.Body.Read(body)
	return string(body)
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var urlCmd tea.Cmd
	var bodyCmd tea.Cmd
	var respCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editing {
			if msg.String() == "esc" {
				m.editing = false
				// TODO(reno): blurring everything seems inefficient
				m.url.Blur()
				m.body.Blur()
			} else {
				// We're editing, put text into a field based on the active section.
				m.url, urlCmd = m.url.Update(msg)
				m.body, bodyCmd = m.body.Update(msg)
			}
		} else {
			switch msg.String() {
			case "1":
				m.activeSection = "URL"
			case "2":
				m.activeSection = "BODY"
			case "3":
				m.activeSection = "RESPONSE"
			case "i":
				m.editing = true
				switch m.activeSection {
				case "URL":
					m.url.Focus()
				case "BODY":
					m.body.Focus()
				}
			case "enter":
				resp := makeRequest(m)
				m.response.SetContent(resp)
				m.r = resp
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		m.flexbox.SetWidth(msg.Width)
		m.flexbox.SetHeight(msg.Height)

	}

	m.response, respCmd = m.response.Update(msg)

	return m, tea.Batch(urlCmd, bodyCmd, respCmd)
}

func (m model) View() string {
	log.Println(m)

	m.flexbox.ForceRecalculate()

	// Method/URL on top
	m.flexbox.Row(0).Cell(0).SetContent(m.method)
	m.flexbox.Row(0).Cell(1).SetContent(m.url.View())

	// Body/response
	m.flexbox.Row(1).Cell(0).SetContent(m.body.View())
	// TODO(reno): Render response in a viewport so we can scroll
	m.flexbox.Row(1).Cell(1).SetContent(m.r)

	// Information: active section & editing
	m.flexbox.Row(2).Cell(0).SetContent("Active section: " + m.activeSection)
	m.flexbox.Row(2).Cell(1).SetContent("Editing?: " + strconv.FormatBool(m.editing))

	return m.flexbox.Render()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())

	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	if _, err := p.Run(); err != nil {
		fmt.Printf("Ahoy me matey, there's been an error: %v", err)
		os.Exit(1)
	}
}
