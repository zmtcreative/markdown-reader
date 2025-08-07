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
    StripH1         bool
    CmdlineOptions  string
    AppProgName         string // Store the application name
    AppProgNameWithExt  string // Store the application name with extension
}

// GetArgs becomes a standalone function that parses flags and returns them.
func GetArgs() (*CliArgs, error) {
    // Use a new, local FlagSet to avoid conflicts with the global one,
    // which is a best practice for robust CLI parsing.
    fs := flag.NewFlagSet("md-reader", flag.ContinueOnError)

    args := &CliArgs{}

    args.ShowHelp = false // Default to not showing help

    // Define flags on the local FlagSet
    var flagNoHTML, flagNoSanitize, flagHelp, flagNoStripH1 bool
    fs.StringVarP(&args.InitialFile, "file", "f", "", "Path to the initial Markdown file")
    fs.BoolVarP(&flagHelp, "help", "h", false, "Show help message")
    fs.BoolVar(&args.StripH1, "strip-h1", true, "Strip the first H1 heading from the Markdown file (default: true)")
    fs.BoolVar(&flagNoStripH1, "nostrip-h1", false, "Do not strip the first H1 heading from the Markdown file (default: strip H1)")
    fs.BoolVar(&flagNoStripH1, "nostrip", false, "Do not strip the first H1 heading from the Markdown file (default: strip H1)")
    fs.BoolVar(&args.AllowInlineHTML, "html", true, "Allow inline HTML rendering (default)")
    fs.BoolVar(&flagNoHTML, "nohtml", false, "Disable inline HTML rendering (default: allow inline HTML)")
    fs.BoolVar(&args.SanitizeHTML, "sanitize", true, "Sanitize HTML and URL output (default)")
    fs.BoolVar(&flagNoSanitize, "nosanitize", false, "Disable sanitizing of HTML and URL output (default: sanitize HTML and URLs)")
    fs.BoolVar(&flagNoSanitize, "nosan", false, "Disable sanitizing of HTML and URL output (default: sanitize HTML and URLs)")

    // Configure flags that can be used without a value
    fs.Lookup("nostrip-h1").NoOptDefVal = "true"
    fs.Lookup("nostrip").NoOptDefVal = "true"
    fs.Lookup("nohtml").NoOptDefVal = "true"
    fs.Lookup("nosanitize").NoOptDefVal = "true"
    fs.Lookup("nosan").NoOptDefVal = "true"
    fs.Lookup("help").NoOptDefVal = "true"
    fs.MarkHidden("nosan") // Hide the alias for nosanitize
    fs.MarkHidden("nostrip") // Hide the alias for nostrip-h1

    // Parse the application's arguments (os.Args[1:])
    // ContinueOnError prevents the flagset from exiting the app on its own.
    // We use _ = fs.Parse(os.Args[1:]) to ignore the error
    _ = fs.Parse(os.Args[1:])

    // Handle non-flag argument as the initial file
    if args.InitialFile == "" && len(fs.Args()) > 0 {
        args.InitialFile = fs.Args()[0]
    }

    if flagNoStripH1 {
        args.StripH1 = false // Disable stripping the first H1 heading
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
    appProgNameWithExt := filepath.Base(os.Args[0])
	appProgName := strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
    usageText.WriteString("<pre>\n")
    usageText.WriteString(fmt.Sprintf("Usage: %s [options] [filepath]\n\n", appProgName))
	usageText.WriteString("  filepath          Path to the Markdown file to open\n")
	usageText.WriteString("                      (if not specified with --file)\n\n")
    usageText.WriteString("Options:\n")
	usageText.WriteString("  -f &lt;path>      Path to the Markdown file to open\n")
    usageText.WriteString("  --file &lt;path>   \n\n")
    usageText.WriteString("  --[no]strip-h1    Enable/Disable stripping the first H1 heading from\n")
    usageText.WriteString("                    the Markdown file\n")
    usageText.WriteString("                      (default: strip first H1 heading and use as Doc Title)\n")
	usageText.WriteString("  --[no]html        Allow/Disallow inline HTML rendering\n")
	usageText.WriteString("                      (default: allow inline HTML)\n")
	usageText.WriteString("  --[no]sanitize    Enable/Disable sanitizing of HTML and URL output\n")
	usageText.WriteString("                      (default: sanitize HTML and URLs)\n")
    usageText.WriteString("</pre>\n")

    args.AppProgName = appProgName
    args.AppProgNameWithExt = appProgNameWithExt
    args.CmdlineOptions = usageText.String()


    // Assign the usage function to the global flag set for the main function to call.
    flag.Usage = func() {
        fmt.Fprint(os.Stderr, args.CmdlineOptions)
    }

    return args, nil
}