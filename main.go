package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Todo struct {
	Text     string  `json:"text"`
	Done     bool    `json:"done"`
	Subtasks []*Todo `json:"subtasks"`
}

type Frame struct {
	art string
}

type Pet struct {
	name   string
	frames []Frame
	color  string
}

// Estilo 8-bit (Pixel Art com Blocos do Terminal)
var pets = []Pet{
	{
		name:  "Cachorro",
		color: "#E8A317", // Laranja/Dourado
		frames: []Frame{
			{art: "  ▄▀▀▀▄  \n ▄█ ▄ █▄ \n ▀█▄▄▄█▀ \n  ▀   ▀  "},
			{art: "  ▄▀▀▀▄  \n ▄█ ▄ █▄ \n ▀█▄▄▄█▀ \n  ▄   ▄  "},
		},
	},
	{
		name:  "Gato",
		color: "#00FFFF", // Ciano
		frames: []Frame{
			{art: " █▀▀▀▀▀█ \n █ ▄ ▄ █ \n ▀▄▄▄▄▄▀ \n  ▀   ▀  "},
			{art: " █▀▀▀▀▀█ \n █ ▀ ▀ █ \n ▀▄▄▄▄▄▀ \n  ▄   ▄  "},
		},
	},
	{
		name:  "Panda",
		color: "#FFFFFF", // Branco
		frames: []Frame{
			{art: " ▄▀▀▀▀▀▄ \n █ █ █ █ \n ▀▄▄▄▄▄▀ \n  ▀   ▀  "},
			{art: " ▄▀▀▀▀▀▄ \n █ █ █ █ \n ▀▄▄▄▄▄▀ \n  ▄   ▄  "},
		},
	},
}

type Config struct {
	KeyBinds string `json:"key_binds"`
	Pet      string `json:"pet"`
	Theme    string `json:"theme"`
}

func loadConfig() Config {
	b, err := os.ReadFile("config.json")
	if err != nil {
		return Config{
			KeyBinds: "normal",
			Pet:      "Gato",
			Theme:    "dark",
		}
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{
			KeyBinds: "normal",
			Pet:      "Gato",
			Theme:    "dark",
		}
	}
	return cfg
}

func saveConfig(cfg Config) {
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err == nil {
		os.WriteFile("config.json", b, 0644)
	}
}

func loadTodos() []*Todo {
	b, err := os.ReadFile("todos.json")
	if err != nil {
		return []*Todo{
			{Text: "Ver os pets andando na barra", Done: true},
			{
				Text: "Limpar casa",
				Done: false,
				Subtasks: []*Todo{
					{Text: "Limpar quarto", Done: false},
					{Text: "Limpar cozinha", Done: false},
				},
			},
			{Text: "Adicionar mais tarefas", Done: false},
		}
	}
	var todos []*Todo
	if err := json.Unmarshal(b, &todos); err != nil {
		return []*Todo{}
	}
	return todos
}

func saveTodos(todos []*Todo) {
	b, _ := json.MarshalIndent(todos, "", "  ")
	os.WriteFile("todos.json", b, 0644)
}

type VisibleTodo struct {
	todo          *Todo
	indent        int
	parent        *Todo
	indexInParent int
}

func getVisible(todos []*Todo) []VisibleTodo {
	var flat []VisibleTodo
	var walk func([]*Todo, int, *Todo)
	walk = func(t []*Todo, indent int, parent *Todo) {
		for i, item := range t {
			flat = append(flat, VisibleTodo{
				todo:          item,
				indent:        indent,
				parent:        parent,
				indexInParent: i,
			})
			walk(item.Subtasks, indent+1, item)
		}
	}
	walk(todos, 0, nil)
	return flat
}

type model struct {
	todos         []*Todo
	cursor        int
	petIndex      int
	frameIndex    int
	position      int
	direction     int
	textInput     textinput.Model
	typing        bool
	addingSubtask bool
	width         int
	height        int
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*250, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Digite a nova tarefa..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	cfg := loadConfig()
	petIdx := 0
	for i, p := range pets {
		if strings.EqualFold(p.name, cfg.Pet) {
			petIdx = i
			break
		}
	}

	return model{
		todos:         loadTodos(),
		cursor:        0,
		petIndex:      petIdx,
		frameIndex:    0,
		position:      0,
		direction:     1,
		textInput:     ti,
		typing:        false,
		addingSubtask: false,
		width:         60, // Fallback width
		height:        24, // Fallback height
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.width < 40 {
			m.width = 40
		}
		maxPos := m.width - 15
		if maxPos < 0 {
			maxPos = 0
		}
		if m.position > maxPos {
			m.position = maxPos
		}
	case tickMsg:
		// Anima os frames do pet
		if len(pets) > 0 && len(pets[m.petIndex].frames) > 0 {
			m.frameIndex = (m.frameIndex + 1) % len(pets[m.petIndex].frames)
		}

		// Move o pet ao longo da tela
		m.position += m.direction * 2

		maxPos := m.width - 15
		if maxPos < 0 {
			maxPos = 0
		}

		// Bate nas bordas e vira
		if m.position >= maxPos { // Limite direito
			m.position = maxPos
			m.direction = -1
		} else if m.position <= 0 { // Limite esquerdo
			m.position = 0
			m.direction = 1
		}

		cmds = append(cmds, tick())

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if !m.typing {
				return m, tea.Quit
			}
		case "up", "k":
			if !m.typing && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			vis := getVisible(m.todos)
			if !m.typing && m.cursor < len(vis)-1 {
				m.cursor++
			}
		case "enter":
			if m.typing {
				if m.textInput.Value() != "" {
					vis := getVisible(m.todos)
					if m.addingSubtask && len(vis) > 0 && m.cursor < len(vis) {
						parentTodo := vis[m.cursor].todo
						parentTodo.Subtasks = append(parentTodo.Subtasks, &Todo{Text: m.textInput.Value(), Done: false})
					} else {
						m.todos = append(m.todos, &Todo{Text: m.textInput.Value(), Done: false})
					}
					m.textInput.SetValue("")
					saveTodos(m.todos)
				}
				m.typing = false
				m.addingSubtask = false
				m.textInput.Blur()
			} else {
				vis := getVisible(m.todos)
				if len(vis) > 0 {
					vis[m.cursor].todo.Done = !vis[m.cursor].todo.Done
					saveTodos(m.todos)
				}
			}
		case "a":
			if !m.typing {
				m.typing = true
				m.addingSubtask = false
				m.textInput.Placeholder = "Digite a nova tarefa..."
				m.textInput.Focus()
				cmds = append(cmds, textinput.Blink)
			}
		case "s":
			if !m.typing && len(getVisible(m.todos)) > 0 {
				m.typing = true
				m.addingSubtask = true
				m.textInput.Placeholder = "Digite a nova sub-tarefa..."
				m.textInput.Focus()
				cmds = append(cmds, textinput.Blink)
			}
		case "d":
			if !m.typing {
				vis := getVisible(m.todos)
				if len(vis) > 0 {
					v := vis[m.cursor]
					if v.parent == nil {
						m.todos = append(m.todos[:v.indexInParent], m.todos[v.indexInParent+1:]...)
					} else {
						v.parent.Subtasks = append(v.parent.Subtasks[:v.indexInParent], v.parent.Subtasks[v.indexInParent+1:]...)
					}
					newVis := getVisible(m.todos)
					if m.cursor >= len(newVis) && m.cursor > 0 {
						m.cursor = len(newVis) - 1
					}
					if m.cursor < 0 {
						m.cursor = 0
					}
					saveTodos(m.todos)
				}
			}
		case "left", "h":
			if !m.typing {
				m.petIndex--
				if m.petIndex < 0 {
					m.petIndex = len(pets) - 1
				}
				m.frameIndex = 0
				cfg := loadConfig()
				cfg.Pet = pets[m.petIndex].name
				saveConfig(cfg)
			}
		case "right", "l":
			if !m.typing {
				m.petIndex++
				if m.petIndex >= len(pets) {
					m.petIndex = 0
				}
				m.frameIndex = 0
				cfg := loadConfig()
				cfg.Pet = pets[m.petIndex].name
				saveConfig(cfg)
			}
		case "esc":
			if m.typing {
				m.typing = false
				m.addingSubtask = false
				m.textInput.Blur()
			}
		}
	}

	if m.typing {
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("#FAFAFA"))

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#EE6FF8")).
				Bold(true)

	doneItemStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("#626262")).
			Strikethrough(true)

	selectedDoneItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#A550A8")).
				Strikethrough(true).
				Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

func (m model) View() string {
	var top strings.Builder

	top.WriteString(titleStyle.Render(" ✓ TUI Todo ") + "\n")

	vis := getVisible(m.todos)

	if len(vis) == 0 {
		top.WriteString(itemStyle.Render("Nenhuma tarefa! Aproveite o dia.") + "\n")
	}

	for i, v := range vis {
		todo := v.todo
		cursor := " "
		if m.cursor == i && !m.typing {
			cursor = ">"
		}

		checked := " "
		if todo.Done {
			checked = "x"
		}

		indentStr := strings.Repeat("  ", v.indent)
		prefix := ""
		if v.indent > 0 {
			prefix = "└─ "
		}

		line := fmt.Sprintf("%s %s%s[%s] %s", cursor, indentStr, prefix, checked, todo.Text)

		if m.cursor == i && !m.typing {
			if todo.Done {
				top.WriteString(selectedDoneItemStyle.Render(line) + "\n")
			} else {
				top.WriteString(selectedItemStyle.Render(line) + "\n")
			}
		} else {
			if todo.Done {
				top.WriteString(doneItemStyle.Render(line) + "\n")
			} else {
				top.WriteString(itemStyle.Render(line) + "\n")
			}
		}
	}

	if m.typing {
		top.WriteString("\n" + m.textInput.View() + "\n")
	}

	var bottom strings.Builder

	// === PETS ANIMADOS 16-BIT ===
	if len(pets) > 0 {
		currentPet := pets[m.petIndex]
		frame := currentPet.frames[m.frameIndex]

		bottom.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#EE6FF8")).Render("❮ "+currentPet.name+" ❯") + "\n")

		// Prepara a animação e o espaçamento para o walk-cycle
		petStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(currentPet.color)).Bold(true)
		lines := strings.Split(frame.art, "\n")
		pos := m.position
		if pos < 0 {
			pos = 0
		}
		padding := strings.Repeat(" ", pos)

		for _, l := range lines {
			bottom.WriteString(padding + petStyle.Render(l) + "\n")
		}
	}

	// Floor bar retro style
	floorChar := "▃"
	floorWidth := m.width - 4
	if floorWidth < 10 {
		floorWidth = 40
	}
	floorBar := lipgloss.NewStyle().Foreground(lipgloss.Color("#4A90E2")).Render(strings.Repeat(floorChar, floorWidth))
	bottom.WriteString(floorBar + "\n")

	// Help
	if m.typing {
		bottom.WriteString(helpStyle.Render("enter: confirmar • esc: cancelar"))
	} else {
		bottom.WriteString(helpStyle.Render("↑/↓: mover • enter: concluir • a: nova • s: sub-tarefa • d: apagar\n←/→: trocar pet • q: sair"))
	}

	topStr := top.String()
	bottomStr := bottom.String()

	topHeight := lipgloss.Height(topStr)
	bottomHeight := lipgloss.Height(bottomStr)
	availableHeight := m.height - 2 // 2 accounts for top and bottom padding of 1
	gap := availableHeight - topHeight - bottomHeight
	if gap < 1 {
		gap = 1
	}

	content := topStr + strings.Repeat("\n", gap) + bottomStr
	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Erro: %v", err)
		os.Exit(1)
	}
}
