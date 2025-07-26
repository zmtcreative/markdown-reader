package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
)

// App struct
type App struct {
    ctx context.Context
    Frontmatter map[string]string // Store frontmatter data here
}

// NewApp creates a new App application struct
func NewApp() *App {
    return &App{
        Frontmatter: map[string]string{},
    }
}

// startup is called when the app starts. The context is created
// after the app is created, but before the event loop starts.
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
}

// GetArgs returns the command line arguments passed to the application.
func (a *App) GetArgs() []string {
    log.Printf("Command line arguments: %s", os.Args[1]) // Log the command line arguments
    return os.Args[1:]
}

// ProcessMarkdown reads a markdown file, renders it to HTML using Goldmark, and returns the HTML string.
func (a *App) ProcessMarkdown(filepath string) (string, error) {
    content, err := os.ReadFile(filepath)
    if err != nil {
        return "", fmt.Errorf("could not read file: %w", err)
    }

    md := goldmark.New(
        goldmark.WithExtensions(
            &frontmatter.Extender{}, // Add the frontmatter extension
        ),
    )

    var buf bytes.Buffer
    var meta map[string]string
    context := parser.NewContext() // Create a context for parsing
    if err := md.Convert(content, &buf, parser.WithContext(context)); err != nil {
        return "", fmt.Errorf("could not convert markdown: %w", err)
    }

    // Extract frontmatter data from the context
    fm := frontmatter.Get(context)
    if fm != nil {
        if err := fm.Decode(&meta); err == nil {
            a.Frontmatter = meta
            log.Printf("Frontmatter data: %v", a.Frontmatter) // Log the frontmatter data
        }
    }

    return buf.String(), nil
}