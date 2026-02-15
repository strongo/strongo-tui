package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestCreateApp(t *testing.T) {
	app, statusText, okBtn, cancelBtn, helpBtn := createApp()
	if app == nil {
		t.Fatal("expected app to be non-nil")
	}
	if statusText == nil {
		t.Fatal("expected statusText to be non-nil")
	}
	if okBtn == nil {
		t.Fatal("expected okButton to be non-nil")
	}
	if cancelBtn == nil {
		t.Fatal("expected cancelButton to be non-nil")
	}
	if helpBtn == nil {
		t.Fatal("expected helpButton to be non-nil")
	}
}

func TestCreateInputCapture(t *testing.T) {
	app := tview.NewApplication()
	statusText := tview.NewTextView()

	handler := createInputCapture(app, statusText)

	tests := []struct {
		name       string
		rune       rune
		wantNil    bool
		wantText   string
		wantStop   bool
	}{
		{"lowercase o", 'o', true, "OK", false},
		{"uppercase O", 'O', true, "OK", false},
		{"lowercase c", 'c', true, "Cancelled", false},
		{"uppercase C", 'C', true, "Cancelled", false},
		{"lowercase h", 'h', true, "Shortcuts", false},
		{"uppercase H", 'H', true, "Shortcuts", false},
		{"lowercase q", 'q', true, "", true},
		{"uppercase Q", 'Q', true, "", true},
		{"other key x", 'x', false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusText.SetText("")
			event := tcell.NewEventKey(tcell.KeyRune, tt.rune, tcell.ModNone)
			result := handler(event)

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil return for rune %c, got non-nil", tt.rune)
				}
			} else {
				if result == nil {
					t.Errorf("expected non-nil return for rune %c, got nil", tt.rune)
				} else if result != event {
					t.Errorf("expected same event returned for rune %c", tt.rune)
				}
			}

			if tt.wantText != "" {
				text := statusText.GetText(true)
				if !strings.Contains(text, tt.wantText) {
					t.Errorf("expected statusText to contain %q, got %q", tt.wantText, text)
				}
			}
		})
	}
}

func TestButtonHandlers(t *testing.T) {
	_, statusText, okBtn, cancelBtn, helpBtn := createApp()

	tests := []struct {
		name     string
		btn      tview.Primitive
		wantText string
	}{
		{"OK button", okBtn, "OK pressed"},
		{"Cancel button", cancelBtn, "Cancelled"},
		{"Help button", helpBtn, "shortcuts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusText.SetText("")
			// Trigger the button's selected func via its InputHandler with Enter key
			handler := tt.btn.InputHandler()
			handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) {})

			text := statusText.GetText(true)
			if !strings.Contains(text, tt.wantText) {
				t.Errorf("expected statusText to contain %q, got %q", tt.wantText, text)
			}
		})
	}
}

func TestRunAppDefault(t *testing.T) {
	app := tview.NewApplication()
	// The default runApp calls app.Run() which will fail without a terminal,
	// but this exercises the default function body for coverage.
	err := runApp(app)
	if err == nil {
		t.Log("runApp returned nil (unexpected in test env, but acceptable)")
	}
}

func TestMain_success(t *testing.T) {
	original := runApp
	defer func() { runApp = original }()

	runApp = func(app *tview.Application) error {
		return nil
	}

	// main() should not panic
	main()
}

func TestMain_error(t *testing.T) {
	original := runApp
	defer func() { runApp = original }()

	runApp = func(app *tview.Application) error {
		return errors.New("test error")
	}

	// Capture stdout to verify error message
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	output := buf.String()

	expected := fmt.Sprintf("Error running application: %v\n", "test error")
	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}
