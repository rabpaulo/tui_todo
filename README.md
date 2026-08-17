# TUI Todo List

A beautiful, retro-styled Terminal User Interface (TUI) Todo List application built in Go using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

## Features

- **Hierarchical Tasks**: Organize your tasks with sub-tasks.
- **Animated 8-Bit Pets**: Cute pixel art companion walking above the floor bar (Dog, Cat, Panda).
- **Interactive UI**: Navigate easily with keyboard shortcuts.
- **Beautiful Styling**: Custom colors and text formatting using Lipgloss.
- **Retro Aesthetic**: Includes a stylized floor bar at the bottom.
- **Docker Support**: Run the application in an isolated container without needing Go installed locally.

## Prerequisites

To run this application natively, you need to have [Go](https://golang.org/doc/install) installed on your system.

Alternatively, you can run it using [Docker](https://docs.docker.com/get-docker/).

## How it Works

The application uses the Bubble Tea architecture (Model, Update, View):
- **Model**: Stores the state of your todos, cursor position, selected pet, and input fields.
- **Update**: Handles keyboard events, timer ticks for pet animation, and updates state.
- **View**: Renders the current state to the terminal using beautiful ANSI styling provided by Lipgloss.

## Installation & Running

### Option 1: Run natively with Go

You can run the application directly using the Go toolchain:

```bash
# Navigate to the project directory
cd /path/to/tui_todo

# Run the application
go run .
```

*Note: A convenience script `todo.sh` is also provided. You can run `./todo.sh` to start the app.*

### Option 2: Run with Docker

If you don't have Go installed or prefer an isolated environment, you can build and run it via Docker:

```bash
# Build the Docker image
docker build -t tui-todo .

# Run the container interactively
docker run -it tui-todo
```

## Keyboard Shortcuts

The application is fully controllable via the keyboard:

| Key | Action |
| --- | --- |
| `↑` or `k` | Move cursor up |
| `↓` or `j` | Move cursor down |
| `enter` | Mark task as done/undone (or confirm when typing) |
| `a` | Add a new task |
| `s` | Add a new sub-task to the currently selected task |
| `d` | Delete the currently selected task or sub-task |
| `←` or `h` | Switch pet (previous) |
| `→` or `l` | Switch pet (next) |
| `esc` | Cancel typing |
| `q` | Quit the application (when not typing) |
| `ctrl+c` | Force quit |
