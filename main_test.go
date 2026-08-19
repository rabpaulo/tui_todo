package main

import (
	"encoding/json"
	"os"
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
		{Text: "Low priority task", Weight: 1},
		{Text: "Medium priority task", Weight: 2},
		{Text: "High priority task", Weight: 3},
	}
	m.cursor = 0
	view := m.View()

	// Ensure explicit debug text is not present
	if strings.Contains(view, "[w:") {
		t.Errorf("expected view not to contain explicit '[w:', got:\n%s", view)
	}

	// Ensure subtle priority indicators are rendered
	if !strings.Contains(view, "!!!") {
		t.Errorf("expected view to contain '!!!' for weight 3, got:\n%s", view)
	}
	if !strings.Contains(view, "!! ") {
		t.Errorf("expected view to contain '!! ' for weight 2, got:\n%s", view)
	}
	if !strings.Contains(view, "!  ") {
		t.Errorf("expected view to contain '!  ' for weight 1, got:\n%s", view)
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

func TestSortTodos(t *testing.T) {
	todos := []*Todo{
		{Text: "Low 1", Weight: 1},
		{Text: "High 1", Weight: 3},
		{
			Text:   "Medium 1",
			Weight: 2,
			Subtasks: []*Todo{
				{Text: "Sub Low", Weight: 1},
				{Text: "Sub High", Weight: 3},
				{Text: "Sub Medium", Weight: 2},
			},
		},
		{Text: "High 2", Weight: 3},
		{Text: "Low 2", Weight: 1},
	}

	sortTodos(todos)

	// Top level checks: High 1 (3), High 2 (3), Medium 1 (2), Low 1 (1), Low 2 (1)
	expectedOrder := []struct {
		text   string
		weight int
	}{
		{"High 1", 3},
		{"High 2", 3},
		{"Medium 1", 2},
		{"Low 1", 1},
		{"Low 2", 1},
	}

	for i, exp := range expectedOrder {
		if todos[i].Text != exp.text || todos[i].Weight != exp.weight {
			t.Errorf("todos[%d] = {%s, %d}, expected {%s, %d}", i, todos[i].Text, todos[i].Weight, exp.text, exp.weight)
		}
	}

	// Subtasks under Medium 1: Sub High (3), Sub Medium (2), Sub Low (1)
	mediumTodo := todos[2]
	expectedSubtasks := []struct {
		text   string
		weight int
	}{
		{"Sub High", 3},
		{"Sub Medium", 2},
		{"Sub Low", 1},
	}

	for i, exp := range expectedSubtasks {
		if mediumTodo.Subtasks[i].Text != exp.text || mediumTodo.Subtasks[i].Weight != exp.weight {
			t.Errorf("mediumTodo.Subtasks[%d] = {%s, %d}, expected {%s, %d}", i, mediumTodo.Subtasks[i].Text, mediumTodo.Subtasks[i].Weight, exp.text, exp.weight)
		}
	}
}

func TestLoadTodosSorting(t *testing.T) {
	origPath := todosFilePath
	tempFile := t.TempDir() + "/todos.json"
	todosFilePath = tempFile
	defer func() { todosFilePath = origPath }()

	content := `[
		{"text": "Task A", "done": false, "weight": 1},
		{"text": "Task B", "done": false, "weight": 3},
		{"text": "Task C", "done": false, "weight": 2}
	]`
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp todos file: %v", err)
	}

	loaded := loadTodos()
	if len(loaded) != 3 {
		t.Fatalf("expected 3 todos, got %d", len(loaded))
	}
	if loaded[0].Text != "Task B" || loaded[0].Weight != 3 {
		t.Errorf("expected loaded[0] to be Task B (weight 3), got %s (%d)", loaded[0].Text, loaded[0].Weight)
	}
	if loaded[1].Text != "Task C" || loaded[1].Weight != 2 {
		t.Errorf("expected loaded[1] to be Task C (weight 2), got %s (%d)", loaded[1].Text, loaded[1].Weight)
	}
	if loaded[2].Text != "Task A" || loaded[2].Weight != 1 {
		t.Errorf("expected loaded[2] to be Task A (weight 1), got %s (%d)", loaded[2].Text, loaded[2].Weight)
	}
}

func TestReorderingOnWeightChangeAndCursorTracking(t *testing.T) {
	origPath := todosFilePath
	todosFilePath = t.TempDir() + "/todos.json"
	defer func() { todosFilePath = origPath }()

	m := initialModel()
	m.todos = []*Todo{
		{Text: "Task 1 (W3)", Weight: 3},
		{Text: "Task 2 (W2)", Weight: 2},
		{Text: "Task 3 (W1)", Weight: 1},
	}
	m.cursor = 2 // on Task 3 (W1)
	m.typing = false

	// Increment weight of Task 3 from 1 to 2 -> should move to index 1 or 2 depending on stable sort (after W2), cursor should track it
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	m = updatedModel.(model)

	if m.todos[m.cursor].Text != "Task 3 (W1)" {
		t.Errorf("expected cursor to track 'Task 3 (W1)', got cursor on '%s'", m.todos[m.cursor].Text)
	}
	if m.todos[m.cursor].Weight != 2 {
		t.Errorf("expected weight 2, got %d", m.todos[m.cursor].Weight)
	}

	// Increment weight of Task 3 from 2 to 3 -> should move above Task 2 (W2), cursor should track it
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	m = updatedModel.(model)

	if m.todos[m.cursor].Text != "Task 3 (W1)" {
		t.Errorf("expected cursor to track 'Task 3 (W1)', got cursor on '%s'", m.todos[m.cursor].Text)
	}
	if m.todos[m.cursor].Weight != 3 {
		t.Errorf("expected weight 3, got %d", m.todos[m.cursor].Weight)
	}
	// It should now be one of the top weight 3 items (index 1)
	if m.cursor != 1 {
		t.Errorf("expected cursor at index 1, got %d", m.cursor)
	}

	// Decrement weight of Task 3 from 3 to 1 with '-' twice
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}})
	m = updatedModel.(model)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}})
	m = updatedModel.(model)

	if m.todos[m.cursor].Text != "Task 3 (W1)" {
		t.Errorf("expected cursor to track 'Task 3 (W1)', got cursor on '%s'", m.todos[m.cursor].Text)
	}
	if m.todos[m.cursor].Weight != 1 {
		t.Errorf("expected weight 1, got %d", m.todos[m.cursor].Weight)
	}
	if m.cursor != 2 {
		t.Errorf("expected cursor at bottom index 2, got %d", m.cursor)
	}
}

func TestTogglePetWithKey(t *testing.T) {
	origConfigPath := configFilePath
	tempConfig := t.TempDir() + "/config.json"
	configFilePath = tempConfig
	defer func() { configFilePath = origConfigPath }()

	origTodosPath := todosFilePath
	todosFilePath = t.TempDir() + "/todos.json"
	defer func() { todosFilePath = origTodosPath }()

	m := initialModel()
	m.width = 60
	m.height = 20
	m.todos = []*Todo{{Text: "Task 1", Weight: 1}}

	if !m.showPet {
		t.Fatalf("expected pet to be visible by default")
	}
	if !strings.Contains(m.View(), "▄▀▀▀▀▀▄") {
		t.Fatalf("expected pet art in View when showPet is true")
	}

	// Press 'p' -> toggle pet off
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updatedModel.(model)

	if m.showPet {
		t.Errorf("expected showPet to be false after pressing 'p'")
	}
	if strings.Contains(m.View(), "▄▀▀▀▀▀▄") {
		t.Errorf("expected pet art to NOT be present when showPet is false")
	}

	// Press 'p' again -> toggle pet back on
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updatedModel.(model)

	if !m.showPet {
		t.Errorf("expected showPet to be true after pressing 'p' again")
	}
	if !strings.Contains(m.View(), "▄▀▀▀▀▀▄") {
		t.Errorf("expected pet art to be present when showPet is toggled back on")
	}

	// In typing mode, pressing 'p' should not toggle pet
	m.typing = true
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updatedModel.(model)
	if !m.showPet {
		t.Errorf("expected showPet to remain true when typing 'p'")
	}
}

func TestPetConfigOption(t *testing.T) {
	origConfigPath := configFilePath
	tempDir := t.TempDir()
	configFilePath = tempDir + "/config.json"
	defer func() { configFilePath = origConfigPath }()

	origTodosPath := todosFilePath
	todosFilePath = tempDir + "/todos.json"
	defer func() { todosFilePath = origTodosPath }()

	// Test 1: Config with show_pet: false
	cfgJSON := `{"pet": "Panda", "show_pet": false}`
	if err := os.WriteFile(configFilePath, []byte(cfgJSON), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	m := initialModel()
	if m.showPet {
		t.Errorf("expected initialModel showPet to be false with show_pet: false in config")
	}

	// Test 2: Config with pet: "none"
	cfgJSON = `{"pet": "none"}`
	if err := os.WriteFile(configFilePath, []byte(cfgJSON), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	m = initialModel()
	if m.showPet {
		t.Errorf("expected initialModel showPet to be false with pet: 'none' in config")
	}

	// Test 3: Config with pet: "off"
	cfgJSON = `{"pet": "off"}`
	if err := os.WriteFile(configFilePath, []byte(cfgJSON), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	m = initialModel()
	if m.showPet {
		t.Errorf("expected initialModel showPet to be false with pet: 'off' in config")
	}

	// Test 4: Config with show_pet: true
	cfgJSON = `{"pet": "Panda", "show_pet": true}`
	if err := os.WriteFile(configFilePath, []byte(cfgJSON), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	m = initialModel()
	if !m.showPet {
		t.Errorf("expected initialModel showPet to be true with show_pet: true in config")
	}
}



