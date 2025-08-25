package app

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// FileManager handles file operations and dialog interactions
type FileManager struct {
    ctx            context.Context
    binaryDetector *BinaryDetector
    docProcessor   *DocumentProcessor
}

// NewFileManager creates a new FileManager
func NewFileManager(ctx context.Context, binaryDetector *BinaryDetector, docProcessor *DocumentProcessor) *FileManager {
    return &FileManager{
        ctx:            ctx,
        binaryDetector: binaryDetector,
        docProcessor:   docProcessor,
    }
}

// OpenFileMenuHandler handles the File > Open menu action
func (fm *FileManager) OpenFileMenuHandler(_ *menu.CallbackData, currentFile *string) {
    log.Println("##> LOG: File -> Open menu item clicked. Opening file dialog...")

    // Open a file dialog to allow the user to select a Markdown file.
    selection, err := runtime.OpenFileDialog(fm.ctx, runtime.OpenDialogOptions{
        Title: "Open Markdown File",
        Filters: []runtime.FileFilter{
            {DisplayName: "Markdown Files (*.md;*.markdown)", Pattern: "*.md;*.markdown"},
            {DisplayName: "All Files (*.*)", Pattern: "*.*"},
        },
    })
    if err != nil {
        if strings.Contains(err.Error(), "The user cancelled the dialog") || strings.Contains(err.Error(), "canceled") {
            log.Println("##> LOG: File dialog cancelled by user.")
            return
        }
        log.Printf("##> LOG: Error opening file dialog: %v", err)
        runtime.EventsEmit(fm.ctx, "error", "Failed to open file dialog: "+err.Error())
        return
    }

    if selection != "" {
        log.Printf("##> LOG: User selected file: %s", selection)

        // Check if the selected file is binary
        isBinary, err := fm.binaryDetector.IsBinaryFile(selection)
        if err != nil {
            log.Printf("##> LOG: Error checking if file is binary %q: %v", selection, err)
            runtime.MessageDialog(fm.ctx, runtime.MessageDialogOptions{
                Type:    runtime.ErrorDialog,
                Title:   "Binary File Check Failed",
                Message: fmt.Sprintf("File MAY be binary: %s\n\nPlease select a text-based Markdown file (.md or .markdown).", filepath.Base(selection)),
            })
            return
        }

        if isBinary {
            log.Printf("##> LOG: User selected a binary file: %s", selection)
            runtime.MessageDialog(fm.ctx, runtime.MessageDialogOptions{
                Type:    runtime.ErrorDialog,
                Title:   "Cannot Open Binary File",
                Message: fmt.Sprintf("Cannot open binary file: %s\n\nPlease select a text-based Markdown file (.md or .markdown).", filepath.Base(selection)),
            })
            return
        }

        err = fm.docProcessor.LoadAndDisplayMarkdown(selection)
        if err != nil {
            log.Printf("##> LOG: Error loading selected Markdown file %q: %v", selection, err)
            runtime.EventsEmit(fm.ctx, "error", "Failed to load selected file: "+err.Error())
        } else {
            log.Printf("##> LOG: Successfully loaded Markdown file: %s", selection)
            *currentFile = selection // Update currentFile to the newly opened file
        }
    } else {
        log.Println("##> LOG: No file selected. User cancelled the dialog.")
    }
}