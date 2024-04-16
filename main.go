package main

import (
	"fmt"
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
var methods = [...]string{"GET", "POST", "PUT", "DELETE"}
var bodyTabs = [...]string{"BODY", "HEADERS"}

type model struct {
	flexbox        *stickers.FlexBox
	envFlexbox     *stickers.FlexBox
	url            textinput.Model
	body           textarea.Model
	bodyTab        int
	headers        [][]string
	headerTable    *stickers.TableSingleType[string]
	selectedHeader int
	method         int
	response       viewport.Model
	r              string
	activeSection  string
	editing        bool
	err            string
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

	envFb := stickers.NewFlexBox(0, 0)

	envR1 := envFb.NewRow().AddCells(
		[]*stickers.FlexBoxCell{
			stickers.NewFlexBoxCell(1, 1),
		},
	)

	envFb.AddRows([]*stickers.FlexBoxRow{envR1})

	headers := [][]string{{"Content-Type", "application/json"}, {"Authorization", "apikey test thing!"}}

	headerTable := stickers.NewTableSingleType[string](0, 0, []string{"Key", "Value"})
	headerTable.SetRatio([]int{1, 1})
	headerTable.AddRows(headers)

	return model{
		flexbox:       fb,
		envFlexbox:    envFb,
		url:           urlInput,
		body:          bodyInput,
		bodyTab:       0,
		headers:       headers,
		headerTable:   headerTable,
		method:        0,
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
	req, err := http.NewRequest(methods[m.method], m.url.Value(), nil)
	check(err)
	for _, h := range m.headers {
		k := h[0]
		v := h[1]
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err.Error()
	}
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
			case "4":
				m.activeSection = "ENVIRONMENT"
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
			case "j":
				switch m.activeSection {
				case "URL":
					m.method = (m.method + 1) % len(methods)
				case "BODY":
					if bodyTabs[m.bodyTab] == "HEADERS" {
						m.headerTable.CursorDown()
					}
				}
			case "k":
				switch m.activeSection {
				case "URL":
					m.method = (m.method - 1 + len(methods)) % len(methods)
				case "BODY":
					if bodyTabs[m.bodyTab] == "HEADERS" {
						m.headerTable.CursorUp()
					}
				}

			case "]":
				switch m.activeSection {
				case "BODY":
					m.bodyTab = (m.bodyTab + 1) % len(bodyTabs)
				}
			case "[":
				switch m.activeSection {
				case "BODY":
					m.bodyTab = (m.bodyTab - 1 + len(bodyTabs)) % len(bodyTabs)
				}
			}

			// TODO(reno): Add a second tab to the body section for headers
			// Also need to figure out a key/value editing system to use for both that and environments
		}
	case tea.WindowSizeMsg:
		m.flexbox.SetWidth(msg.Width)
		m.flexbox.SetHeight(msg.Height)
		m.headerTable.SetWidth(msg.Width / 2)
		m.headerTable.SetHeight(msg.Height / 2)
	}

	m.response, respCmd = m.response.Update(msg)

	return m, tea.Batch(urlCmd, bodyCmd, respCmd)
}

func PrintHeaders(h map[string]string) string {
	// finalStr := ""
	// for k, v := range h {
	//
	// }
	return ""
}

func (m model) View() string {
	// DEBUG: Print out the model. This is sort of useless.
	// log.Println(m)
	if m.activeSection == "ENVIRONMENT" {
		m.envFlexbox.ForceRecalculate()

		m.envFlexbox.Row(0).Cell(0).SetContent("Test envinronments!")

		return m.envFlexbox.Render()
	}

	m.flexbox.ForceRecalculate()

	// Method/URL on top
	m.flexbox.Row(0).Cell(0).SetContent(methods[m.method])
	m.flexbox.Row(0).Cell(1).SetContent(m.url.View())

	// Body/response
	bodyContent := ""
	switch bodyTabs[m.bodyTab] {
	case "BODY":
		bodyContent = m.body.View()
	case "HEADERS":
		bodyContent = m.headerTable.Render()
	}

	m.flexbox.Row(1).Cell(0).SetContent(bodyContent)
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
