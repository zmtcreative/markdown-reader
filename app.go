package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/yuin/goldmark"
)

// App struct
type App struct {
    ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
    return &App{}
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
        goldmark.WithExtensions(),
    )

    var buf strings.Builder
    if err := md.Convert(content, &buf); err != nil {
        return "", fmt.Errorf("could not convert markdown: %w", err)
    }

    return buf.String(), nil
}