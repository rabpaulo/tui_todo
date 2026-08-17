package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Todo struct {
	text string
	done bool
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
		name: "Cachorro",
		color: "#E8A317", // Laranja/Dourado
		frames: []Frame{
			{art: "  ▄▀▀▀▄  \n ▄█ ▄ █▄ \n ▀█▄▄▄█▀ \n  ▀   ▀  "},
			{art: "  ▄▀▀▀▄  \n ▄█ ▄ █▄ \n ▀█▄▄▄█▀ \n  ▄   ▄  "},
		},
	},
	{
		name: "Gato",
		color: "#00FFFF", // Ciano
		frames: []Frame{
			{art: " █▀▀▀▀▀█ \n █ ▄ ▄ █ \n ▀▄▄▄▄▄▀ \n  ▀   ▀  "},
			{art: " █▀▀▀▀▀█ \n █ ▀ ▀ █ \n ▀▄▄▄▄▄▀ \n  ▄   ▄  "},
		},
	},
	{
		name: "Panda",
		color: "#FFFFFF", // Branco
		frames: []Frame{
			{art: " ▄▀▀▀▀▀▄ \n █ █ █ █ \n ▀▄▄▄▄▄▀ \n  ▀   ▀  "},
			{art: " ▄▀▀▀▀▀▄ \n █ █ █ █ \n ▀▄▄▄▄▄▀ \n  ▄   ▄  "},
		},
	},
}

type model struct {
	todos      []Todo
	cursor     int
	petIndex   int
	frameIndex int
	position   int
	direction  int
	textInput  textinput.Model
	typing     bool
	width      int
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

	return model{
		todos: []Todo{
			{text: "Ver os pets andando na barra", done: true},
			{text: "Adicionar mais tarefas", done: false},
		},
		cursor:     0,
		petIndex:   0,
		frameIndex: 0,
		position:   0,
		direction:  1,
		textInput:  ti,
		typing:     false,
		width:      60, // Fallback width
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
		if m.width < 40 {
			m.width = 40
		}
	case tickMsg:
		// Anima os frames do pet
		m.frameIndex = (m.frameIndex + 1) % len(pets[m.petIndex].frames)
		
		// Move o pet ao longo da tela
		m.position += m.direction * 2
		
		// Bate nas bordas e vira
		if m.position >= m.width-15 { // Limite direito
			m.direction = -1
		} else if m.position <= 0 { // Limite esquerdo
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
			if !m.typing && m.cursor < len(m.todos)-1 {
				m.cursor++
			}
		case "enter":
			if m.typing {
				if m.textInput.Value() != "" {
					m.todos = append(m.todos, Todo{text: m.textInput.Value(), done: false})
					m.textInput.SetValue("")
				}
				m.typing = false
				m.textInput.Blur()
			} else {
				if len(m.todos) > 0 {
					m.todos[m.cursor].done = !m.todos[m.cursor].done
				}
			}
		case "a":
			if !m.typing {
				m.typing = true
				m.textInput.Focus()
				cmds = append(cmds, textinput.Blink)
			}
		case "d":
			if !m.typing && len(m.todos) > 0 {
				m.todos = append(m.todos[:m.cursor], m.todos[m.cursor+1:]...)
				if m.cursor >= len(m.todos) && m.cursor > 0 {
					m.cursor--
				}
			}
		case "left", "h":
			if !m.typing {
				m.petIndex--
				if m.petIndex < 0 {
					m.petIndex = len(pets) - 1
				}
				m.frameIndex = 0
			}
		case "right", "l":
			if !m.typing {
				m.petIndex++
				if m.petIndex >= len(pets) {
					m.petIndex = 0
				}
				m.frameIndex = 0
			}
		case "esc":
			if m.typing {
				m.typing = false
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
	var b strings.Builder

	b.WriteString(titleStyle.Render(" ✓ TUI Todo ") + "\n")

	if len(m.todos) == 0 {
		b.WriteString(itemStyle.Render("Nenhuma tarefa! Aproveite o dia.") + "\n\n")
	}

	for i, todo := range m.todos {
		cursor := " "
		if m.cursor == i && !m.typing {
			cursor = ">"
		}

		checked := " "
		if todo.done {
			checked = "x"
		}

		line := fmt.Sprintf("%s [%s] %s", cursor, checked, todo.text)
		
		if m.cursor == i && !m.typing {
			if todo.done {
				b.WriteString(selectedDoneItemStyle.Render(line) + "\n")
			} else {
				b.WriteString(selectedItemStyle.Render(line) + "\n")
			}
		} else {
			if todo.done {
				b.WriteString(doneItemStyle.Render(line) + "\n")
			} else {
				b.WriteString(itemStyle.Render(line) + "\n")
			}
		}
	}

	if m.typing {
		b.WriteString("\n" + m.textInput.View() + "\n")
	} else {
		b.WriteString("\n")
	}

	// === PETS ANIMADOS 16-BIT ===
	currentPet := pets[m.petIndex]
	frame := currentPet.frames[m.frameIndex]
	
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#EE6FF8")).Render("❮ " + currentPet.name + " ❯") + "\n")

	// Prepara a animação e o espaçamento para o walk-cycle
	petStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(currentPet.color)).Bold(true)
	lines := strings.Split(frame.art, "\n")
	padding := strings.Repeat(" ", m.position)
	
	for _, l := range lines {
		b.WriteString(padding + petStyle.Render(l) + "\n")
	}

	// Floor bar retro style
	floorChar := "▃"
	floorWidth := m.width - 4
	if floorWidth < 10 {
		floorWidth = 40
	}
	floorBar := lipgloss.NewStyle().Foreground(lipgloss.Color("#4A90E2")).Render(strings.Repeat(floorChar, floorWidth))
	b.WriteString(floorBar + "\n")

	// Help
	if m.typing {
		b.WriteString(helpStyle.Render("enter: confirmar • esc: cancelar"))
	} else {
		b.WriteString(helpStyle.Render("↑/↓: mover • enter: concluir • a: adicionar • d: apagar\n←/→: trocar pet • q: sair"))
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Erro: %v", err)
		os.Exit(1)
	}
}
