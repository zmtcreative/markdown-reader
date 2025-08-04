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
    fs := flag.NewFlagSet("md-reader", flag.ContinueOnError)

    args := &CliArgs{}

    args.ShowHelp = false // Default to not showing help

    // Define flags on the local FlagSet
    var flagNoHTML, flagNoSanitize, flagHelp bool
    fs.StringVarP(&args.InitialFile, "file", "f", "", "Path to the initial Markdown file")
    fs.BoolVarP(&flagHelp, "help", "h", false, "Show help message")
    fs.BoolVar(&args.AllowInlineHTML, "html", true, "Allow inline HTML rendering (default)")
    fs.BoolVar(&flagNoHTML, "nohtml", false, "Disable inline HTML rendering (default: allow inline HTML)")
    fs.BoolVar(&args.SanitizeHTML, "sanitize", true, "Sanitize HTML and URL output (default)")
    fs.BoolVar(&flagNoSanitize, "nosanitize", false, "Disable sanitizing of HTML and URL output (default: sanitize HTML and URLs)")
    fs.BoolVar(&flagNoSanitize, "nosan", false, "Disable sanitizing of HTML and URL output (default: sanitize HTML and URLs)")

    // Configure flags that can be used without a value
    fs.Lookup("nohtml").NoOptDefVal = "true"
    fs.Lookup("nosanitize").NoOptDefVal = "true"
    fs.Lookup("nosan").NoOptDefVal = "true"
    fs.Lookup("help").NoOptDefVal = "true"
    fs.MarkHidden("nosan") // Hide the alias for nosanitize

    // Parse the application's arguments (os.Args[1:])
    // ContinueOnError prevents the flagset from exiting the app on its own.
    // We use _ = fs.Parse(os.Args[1:]) to ignore the error
    _ = fs.Parse(os.Args[1:])

    // Handle non-flag argument as the initial file
    if args.InitialFile == "" && len(fs.Args()) > 0 {
        args.InitialFile = fs.Args()[0]
    }

    if flagNoHTML {
        args.AllowInlineHTML = false // Disable inline HTML rendering
    }
    if flagNoSanitize {
        args.SanitizeHTML = false // Disable sanitizing of HTML and URLs
    }
    if flagHelp {
        args.ShowHelp = true // Set the flag to show help
    }

    // Build the usage string
    var usageText strings.Builder
    appNameWithExt := filepath.Base(os.Args[0])
	appName := strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
    usageText.WriteString("<pre>\n")
    usageText.WriteString(fmt.Sprintf("Usage: %s [options] [filepath]\n\n", appName))
	usageText.WriteString("  filepath        Path to the Markdown file to open\n")
	usageText.WriteString("                    (if not specified with --file)\n\n")
    usageText.WriteString("Options:\n")
	usageText.WriteString("  -f &lt;path>       Path to the Markdown file to open\n")
    usageText.WriteString("  --file &lt;path>   \n\n")
	usageText.WriteString("  --nohtml        Disable inline HTML rendering\n")
	usageText.WriteString("                    (default: allow inline HTML)\n")
	usageText.WriteString("  --nosanitize    Disable sanitizing of HTML and URL output\n")
	usageText.WriteString("                    (default: sanitize HTML and URLs)\n")
    usageText.WriteString("</pre>\n")

    args.AppName = appName
    args.AppNameWithExt = appNameWithExt
    args.CmdlineOptions = usageText.String()


    // Assign the usage function to the global flag set for the main function to call.
    flag.Usage = func() {
        fmt.Fprint(os.Stderr, args.CmdlineOptions)
    }

    return args, nil
}