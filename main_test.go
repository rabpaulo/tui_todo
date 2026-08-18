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
