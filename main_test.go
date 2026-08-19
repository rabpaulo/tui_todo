package main

import (
	"encoding/json"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNormalizeTodos(t *testing.T) {
	todos := []*Todo{
		{Text: "T1", Weight: 0, Subtasks: []*Todo{
			{Text: "S1", Weight: -5},
			{Text: "S2", Weight: 10},
			{Text: "S3", Weight: 2},
		}},
		{Text: "T2", Weight: 4},
		{Text: "T3", Weight: 1},
	}

	normalizeTodos(todos)

	if todos[0].Weight != 1 {
		t.Errorf("expected weight 1, got %d", todos[0].Weight)
	}
	if todos[0].Subtasks[0].Weight != 1 {
		t.Errorf("expected weight 1 for S1, got %d", todos[0].Subtasks[0].Weight)
	}
	if todos[0].Subtasks[1].Weight != 3 {
		t.Errorf("expected weight 3 for S2, got %d", todos[0].Subtasks[1].Weight)
	}
	if todos[0].Subtasks[2].Weight != 2 {
		t.Errorf("expected weight 2 for S3, got %d", todos[0].Subtasks[2].Weight)
	}
	if todos[1].Weight != 3 {
		t.Errorf("expected weight 3 for T2, got %d", todos[1].Weight)
	}
	if todos[2].Weight != 1 {
		t.Errorf("expected weight 1 for T3, got %d", todos[2].Weight)
	}
}

func TestWeightIncrementDecrement(t *testing.T) {
	origPath := todosFilePath
	todosFilePath = t.TempDir() + "/todos.json"
	defer func() { todosFilePath = origPath }()

	m := initialModel()
	m.todos = []*Todo{
		{Text: "Task 1", Weight: 2},
	}
	m.cursor = 0
	m.typing = false

	// Increment (=) -> should be 3
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'='}})
	m = updatedModel.(model)
	if m.todos[0].Weight != 3 {
		t.Errorf("expected weight 3 after '=', got %d", m.todos[0].Weight)
	}

	// Increment (+) beyond max -> should remain 3
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	m = updatedModel.(model)
	if m.todos[0].Weight != 3 {
		t.Errorf("expected weight 3 capped at 3, got %d", m.todos[0].Weight)
	}

	// Decrement (-) -> should be 2
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}})
	m = updatedModel.(model)
	if m.todos[0].Weight != 2 {
		t.Errorf("expected weight 2 after '-', got %d", m.todos[0].Weight)
	}

	// Decrement (-) -> should be 1
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}})
	m = updatedModel.(model)
	if m.todos[0].Weight != 1 {
		t.Errorf("expected weight 1 after '-', got %d", m.todos[0].Weight)
	}

	// Decrement (-) below min -> should remain 1
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}})
	m = updatedModel.(model)
	if m.todos[0].Weight != 1 {
		t.Errorf("expected weight 1 floored at 1, got %d", m.todos[0].Weight)
	}
}

func TestTodoJSONSerialization(t *testing.T) {
	todo := &Todo{
		Text:   "Test task",
		Done:   false,
		Weight: 3,
	}

	data, err := json.Marshal(todo)
	if err != nil {
		t.Fatalf("failed to marshal todo: %v", err)
	}

	var unmarshaled Todo
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal todo: %v", err)
	}

	if unmarshaled.Weight != 3 {
		t.Errorf("expected unmarshaled weight 3, got %d", unmarshaled.Weight)
	}
}

func TestViewContainsWeight(t *testing.T) {
	m := initialModel()
	m.todos = []*Todo{
		{Text: "High priority task", Weight: 3},
	}
	m.cursor = 0
	view := m.View()

	if !strings.Contains(view, "[w:3]") {
		t.Errorf("expected view to contain '[w:3]', got:\n%s", view)
	}
}

func TestTerminalResizeWindowSizeMsg(t *testing.T) {
	m := initialModel()

	// Resize to wide terminal
	updatedModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updatedModel.(model)

	if m.width != 120 || m.height != 40 {
		t.Errorf("expected dimensions 120x40, got %dx%d", m.width, m.height)
	}
	if m.contentWidth() != 116 {
		t.Errorf("expected contentWidth 116, got %d", m.contentWidth())
	}
	if m.textInput.Width != 112 {
		t.Errorf("expected textInput width 112, got %d", m.textInput.Width)
	}

	// Resize to narrow terminal
	updatedModel, _ = m.Update(tea.WindowSizeMsg{Width: 28, Height: 10})
	m = updatedModel.(model)

	if m.width != 28 || m.height != 10 {
		t.Errorf("expected dimensions 28x10, got %dx%d", m.width, m.height)
	}
	if m.contentWidth() != 28 {
		t.Errorf("expected contentWidth 28, got %d", m.contentWidth())
	}
	if m.textInput.Width != 24 {
		t.Errorf("expected textInput width 24, got %d", m.textInput.Width)
	}
}

func TestResponsiveTextWrapping(t *testing.T) {
	m := initialModel()
	longText := "This is an extremely long todo item description that should be wrapped properly according to the terminal width without causing overflow errors"
	m.todos = []*Todo{
		{Text: longText, Weight: 1},
	}
	m.cursor = 0

	// Test with narrow terminal width
	m.width = 40
	m.height = 24
	view := m.View()

	if !strings.Contains(view, "This is an extremely") {
		t.Errorf("expected view to contain start of long text, got:\n%s", view)
	}
}

func TestListScrollingOnSmallTerminal(t *testing.T) {
	m := initialModel()
	m.todos = []*Todo{
		{Text: "Task 0", Weight: 1},
		{Text: "Task 1", Weight: 1},
		{Text: "Task 2", Weight: 1},
		{Text: "Task 3", Weight: 1},
		{Text: "Task 4", Weight: 1},
		{Text: "Task 5", Weight: 1},
		{Text: "Task 6", Weight: 1},
		{Text: "Task 7", Weight: 1},
	}
	m.width = 60
	m.height = 14 // Constrained height

	// Cursor at top
	m.cursor = 0
	view := m.View()
	if !strings.Contains(view, "Task 0") {
		t.Errorf("expected view with cursor=0 to contain Task 0, got:\n%s", view)
	}
	if !strings.Contains(view, "below") {
		t.Errorf("expected view with cursor=0 to indicate more tasks below, got:\n%s", view)
	}

	// Move cursor to bottom
	m.cursor = 7
	view = m.View()
	if !strings.Contains(view, "Task 7") {
		t.Errorf("expected view with cursor=7 to contain Task 7, got:\n%s", view)
	}
	if !strings.Contains(view, "above") {
		t.Errorf("expected view with cursor=7 to indicate more tasks above, got:\n%s", view)
	}
}

func TestPetVisibilityOnTerminalResize(t *testing.T) {
	m := initialModel()
	m.todos = []*Todo{
		{Text: "Task 1", Weight: 1},
	}

	// Height >= 14: Pet should be visible
	m.width = 60
	m.height = 20
	view := m.View()
	if !strings.Contains(view, "▄▀▀▀▀▀▄") {
		t.Errorf("expected pet art to be rendered when height=20, got:\n%s", view)
	}

	// Height < 14: Pet should be hidden to conserve space for tasks
	m.height = 10
	view = m.View()
	if strings.Contains(view, "▄▀▀▀▀▀▄") {
		t.Errorf("expected pet art to be hidden when height=10, got:\n%s", view)
	}
	// Tasks and help must still be visible
	if !strings.Contains(view, "Task 1") {
		t.Errorf("expected task to be visible even with height=10, got:\n%s", view)
	}
}

