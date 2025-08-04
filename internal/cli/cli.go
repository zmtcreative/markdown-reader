package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
)

// Define a struct to hold the parsed command-line arguments
type CliArgs struct {
    InitialFile     string
    AllowInlineHTML bool
    SanitizeHTML    bool
    ShowHelp        bool
    CmdlineOptions  string
    AppName         string // Store the application name
    AppNameWithExt  string // Store the application name with extension
}

// GetArgs becomes a standalone function that parses flags and returns them.
func GetArgs() (*CliArgs, error) {
    // Use a new, local FlagSet to avoid conflicts with the global one,
    // which is a best practice for robust CLI parsing.
    fs := flag.NewFlagSet("markdown-reader", flag.ContinueOnError)

    args := &CliArgs{}

    // Define flags on the local FlagSet
    fs.StringVarP(&args.InitialFile, "file", "f", "", "Path to the initial Markdown file")
    fs.BoolVar(&args.AllowInlineHTML, "nohtml", true, "No inline HTML rendering (default: allow inline HTML)")
    fs.BoolVar(&args.SanitizeHTML, "nosanitize", true, "No sanitizing of HTML and URL output (default: sanitize HTML and URLs)")
    // fs.BoolVarP(&args.ShowHelp, "help", "h", false, "Display help message")

    // Configure flags that can be used without a value
    fs.Lookup("nohtml").NoOptDefVal = "false"
    fs.Lookup("nosanitize").NoOptDefVal = "false"

    // Parse the application's arguments (os.Args[1:])
    // ContinueOnError prevents the flagset from exiting the app on its own.
    err := fs.Parse(os.Args[1:])
    if err != nil {
        // If parsing fails for any reason (e.g., bad flag), return the error.
        return nil, err
    }

    // Handle non-flag argument as the initial file
    if args.InitialFile == "" && len(fs.Args()) > 0 {
        args.InitialFile = fs.Args()[0]
    }

    // Build the usage string
    var usageText strings.Builder
    appNameWithExt := filepath.Base(os.Args[0])
	appName := strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
    usageText.WriteString(fmt.Sprintf("<pre>\nUsage: %s [options] [filepath]\n\n", appName))
	usageText.WriteString("  filepath             Path to the initial Markdown file\n")
	usageText.WriteString("                         (if not specified with --file)\n\n")
    usageText.WriteString("Options:\n")
	usageText.WriteString("  -f, --file=<path>    Path to the initial Markdown file\n")
	usageText.WriteString("      --nohtml         Disable inline HTML rendering\n")
	usageText.WriteString("                         (default: allow inline HTML)\n")
	usageText.WriteString("      --nosanitize     Disable sanitizing of HTML and URL output\n")
	usageText.WriteString("                         (default: sanitize HTML and URLs)\n</pre>\n")

    args.AppName = appName
    args.AppNameWithExt = appNameWithExt
    args.CmdlineOptions = usageText.String()


    // Assign the usage function to the global flag set for the main function to call.
    flag.Usage = func() {
        fmt.Fprint(os.Stderr, args.CmdlineOptions)
    }

    return args, nil
}