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
    InitialFile         *string
    ShowHelp            *bool
    CmdlineOptions      *string
    AppProgName         *string // Store the application name
    AppProgNameWithExt  *string // Store the application name with extension
}

// GetArgs becomes a standalone function that parses flags and returns them.
func GetArgs() (*CliArgs, error) {
    // Use a new, local FlagSet to avoid conflicts with the global one,
    // which is a best practice for robust CLI parsing.
    fs := flag.NewFlagSet("md-reader", flag.ContinueOnError)

    args := &CliArgs{}

    // Initialize with nil pointers - this indicates no CLI override was provided
    args.ShowHelp = nil
    args.InitialFile = nil

    // Define local variables for flag parsing
    var initialFile string
    var flagHelp bool

    // Define flags on the local FlagSet
    fs.StringVarP(&initialFile, "file", "f", "", "Path to the initial Markdown file")
    fs.BoolVarP(&flagHelp, "help", "h", false, "Show help message")

    // Configure flags that can be used without a value
    fs.Lookup("help").NoOptDefVal = "true"

    // Parse the application's arguments (os.Args[1:])
    // ContinueOnError prevents the flagset from exiting the app on its own.
    // We use _ = fs.Parse(os.Args[1:]) to ignore the error
    if len(os.Args) > 1 {
        _ = fs.Parse(os.Args[1:])
    } else {
        _ = fs.Parse([]string{})
    }

    // Handle non-flag argument as the initial file (positional args take precedence)
    if len(fs.Args()) > 0 {
        initialFile = fs.Args()[0]
    }

    // Set values only if they were explicitly provided or inferred
    if initialFile != "" {
        args.InitialFile = &initialFile
    }

    if flagHelp {
        args.ShowHelp = &flagHelp
    }

    // Build the usage string
    var usageText strings.Builder
    var appProgNameWithExt, appProgName string

    if len(os.Args) > 0 {
        appProgNameWithExt = filepath.Base(os.Args[0])
        appProgName = strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
    } else {
        appProgNameWithExt = "md-reader"
        appProgName = "md-reader"
    }

    usageText.WriteString("<pre>\n")
    usageText.WriteString(fmt.Sprintf("Usage: %s [options] [filepath]\n\n", appProgName))
	usageText.WriteString("  filepath          Path to the Markdown file to open\n")
	usageText.WriteString("                      (if not specified with --file)\n\n")
    usageText.WriteString("Options:\n")
	usageText.WriteString("  -f &lt;path&gt;      Path to the Markdown file to open\n")
    usageText.WriteString("  --file &lt;path&gt;   \n\n")
    usageText.WriteString("</pre>\n")

    cmdlineOptions := usageText.String()
    args.AppProgName = &appProgName
    args.AppProgNameWithExt = &appProgNameWithExt
    args.CmdlineOptions = &cmdlineOptions


    // Assign the usage function to the global flag set for the main function to call.
    flag.Usage = func() {
        if args.CmdlineOptions != nil {
            fmt.Fprint(os.Stderr, *args.CmdlineOptions)
        }
    }

    return args, nil
}