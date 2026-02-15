# strongo-tui

[![Go CI](https://github.com/strongo/strongo-tui/actions/workflows/ci.yml/badge.svg)](https://github.com/strongo/strongo-tui/actions/workflows/ci.yml)
[![Coverage Status](https://coveralls.io/repos/github/strongo/strongo-tui/badge.svg?branch=main)](https://coveralls.io/github/strongo/strongo-tui?branch=main)

Reusable terminal UI components for building TUI applications with Go, powered by [tview](https://github.com/rivo/tview).

## Overview

`strongo-tui` is a collection of common terminal UI building blocks extracted from production applications like [datatug-cli](https://github.com/datatug/datatug-cli) and [filetug](https://github.com/filetug/filetug). It provides ready-to-use components and utilities for building modern, interactive terminal user interfaces in Go.

## Features

- **Colors & Theming**: ANSI color helpers and tcell-based theme support
- **Reusable Components**: Buttons, boxes, padding utilities
- **Consistent Styling**: Default border styles with focus/blur states
- **Type-safe**: Uses Go generics for flexible component wrapping
- **Well-tested**: Comprehensive test coverage

## Installation

```bash
go get github.com/strongo/strongo-tui
```

## Components

### Colors

The `colors` package provides both ANSI escape codes for terminal output and tcell colors for tview components.

#### ANSI Colors

```go
import "github.com/strongo/strongo-tui/pkg/colors"

// Basic colors
fmt.Println(colors.RedText("Error message"))
fmt.Println(colors.GreenText("Success!"))
fmt.Println(colors.YellowText("Warning"))
fmt.Println(colors.BlueText("Info"))
fmt.Println(colors.GrayText("Debug"))

// Semantic colors
fmt.Println(colors.Danger("Critical error"))
fmt.Println(colors.Warning("Be careful"))
fmt.Println(colors.Success("Operation completed"))
```

#### TCell Colors

```go
import "github.com/strongo/strongo-tui/pkg/colors"

// Pre-defined colors for UI components
box.SetBorderColor(colors.DefaultFocusedBorderColor)
table.SetColumnColor(colors.TableColumnTitle)
```

### Themes

The `themes` package provides consistent theming for terminal UI components.

```go
import (
    "github.com/strongo/strongo-tui/pkg/themes"
    "github.com/rivo/tview"
)

// Apply default border with padding
box := tview.NewBox()
themes.DefaultBorderWithPadding(box)

// Apply default border without padding
themes.DefaultBorderWithoutPadding(box)

// Set panel title with styling
themes.SetPanelTitle(box, "My Panel")

// Access current theme
theme := themes.CurrentTheme
color := theme.FocusedBorderColor
```

### Boxed Components

The `boxed` package provides utilities for wrapping tview primitives with boxes and borders.

```go
import (
    "github.com/strongo/strongo-tui/pkg/components/boxed"
    "github.com/rivo/tview"
)

// Wrap a primitive with default borders
textView := tview.NewTextView()
box := tview.NewBox()
wrapped := boxed.WithDefaultBorders(textView, box)

// Access the original primitive
original := wrapped.GetPrimitive()

// Access the box for customization
theBox := wrapped.GetBox()
```

### Button with Shortcut

Enhanced button component that displays keyboard shortcuts.

```go
import "github.com/strongo/strongo-tui/pkg/components/button"

// Create button with shortcut
saveButton := button.NewWithShortcut("Save", 's')
// Displays as: "(s) Save"

// Customize shortcut style
saveButton.SetShortcutStyle(
    tcell.StyleDefault.
        Foreground(tcell.ColorYellow).
        Background(tcell.ColorBlack),
)
```

### Padded Box

Create boxes with real layout padding (not just border padding).

```go
import "github.com/strongo/strongo-tui/pkg/components/padding"

content := tview.NewTextView().SetText("Hello, World!")
padded := padding.Box(
    content,
    "My Title",
    1, // top padding
    1, // bottom padding
    2, // left padding
    2, // right padding
)
```

## Example Application

```go
package main

import (
    "github.com/rivo/tview"
    "github.com/strongo/strongo-tui/pkg/colors"
    "github.com/strongo/strongo-tui/pkg/components/button"
    "github.com/strongo/strongo-tui/pkg/themes"
)

func main() {
    app := tview.NewApplication()

    // Create a text view with some content
    textView := tview.NewTextView().
        SetText(colors.Success("Welcome to strongo-tui!") + "\n\n" +
            "This is a sample application demonstrating the components.")
    
    // Apply themed border
    themes.SetPanelTitle(textView.Box, "Demo Application")

    // Create buttons with shortcuts
    okButton := button.NewWithShortcut("OK", 'o')
    cancelButton := button.NewWithShortcut("Cancel", 'c')

    // Create button bar
    buttons := tview.NewFlex().
        AddItem(okButton, 10, 0, true).
        AddItem(nil, 2, 0, false).
        AddItem(cancelButton, 14, 0, false)

    // Main layout
    flex := tview.NewFlex().
        SetDirection(tview.FlexRow).
        AddItem(textView, 0, 1, false).
        AddItem(buttons, 3, 0, true)

    if err := app.SetRoot(flex, true).Run(); err != nil {
        panic(err)
    }
}
```

## Dependencies

- [github.com/rivo/tview](https://github.com/rivo/tview) v0.42.0 - Rich terminal UI library
- [github.com/gdamore/tcell/v2](https://github.com/gdamore/tcell) v2.13.8 - Terminal handling
- [github.com/alecthomas/chroma/v2](https://github.com/alecthomas/chroma) v2.23.1 - Syntax highlighting

## Development

### Building

```bash
go build ./...
```

### Testing

```bash
go test ./...
```

### Running Examples

```bash
cd examples/demo
go run main.go
```

## Origin

This library extracts and consolidates common terminal UI code from:

- [datatug/datatug-cli](https://github.com/datatug/datatug-cli) - CLI-first data exploration platform
- [filetug/filetug](https://github.com/filetug/filetug) - Modern CLI file browser

By extracting these components into a shared library, we make it easier to build consistent terminal UIs across multiple projects.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

Apache License 2.0 - See LICENSE file for details.

## Related Projects

- [datatug-cli](https://github.com/datatug/datatug-cli) - Cross-database data exploration tool
- [filetug](https://github.com/filetug/filetug) - Modern terminal file browser
- [tview](https://github.com/rivo/tview) - Rich terminal UI library for Go
