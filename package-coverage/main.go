package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/corsc/go-tools/package-coverage/generator"
	"github.com/corsc/go-tools/package-coverage/parser"
	"github.com/corsc/go-tools/package-coverage/utils"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Error: %s\n", r)
		}
	}()

	verbose := false
	quiet := true
	coverage := false
	singleDir := false
	doClean := false
	doPrint := false
	ignorePaths := ""
	webHook := ""
	channelOverride := ""
	prefix := ""
	depth := 0
	minCoverage := 0
	var exclusions *regexp.Regexp

	flag.BoolVar(&verbose, "v", false, "verbose mode is useful for debugging this tool")
	flag.BoolVar(&quiet, "q", true, "quiet mode will suppress the stdOut messages from go test")
	flag.BoolVar(&coverage, "c", false, "generate coverage")
	flag.BoolVar(&singleDir, "s", false, "only generate for the supplied directory (no recursion / will ignore -i)")
	flag.BoolVar(&doClean, "d", false, "clean")
	flag.BoolVar(&doPrint, "p", false, "print coverage to stdout")
	flag.StringVar(&ignorePaths, "i", `./\.git.*|./_.*`, "ignore file paths matching the specified regex (match directories by surrounding the directory name with slashes; match files by prefixing with a slash)")
	flag.StringVar(&webHook, "webhook", "", "Slack webhook URL (missing means don't send)")
	flag.StringVar(&channelOverride, "channel", "", "Slack channel (missing means use the default channel for this webhook)")
	flag.StringVar(&prefix, "prefix", "", "Prefix to be removed from the output")
	flag.IntVar(&depth, "depth", 0, "How many levels of coverage to output (default is 0 = all)")
	flag.IntVar(&minCoverage, "m", 0, "minimum coverage")
	flag.Parse()

	startDir := utils.GetCurrentDir()
	path := getPath()
	goTestArgs := getGoTestArguments()

	if verbose {
		utils.LogWhenVerbose("Config:")
		utils.LogWhenVerbose("Verbose: %v", verbose)
		utils.LogWhenVerbose("Quiet: %v", quiet)
		utils.LogWhenVerbose("Generate Coverage: %v", coverage)
		utils.LogWhenVerbose("Single Directory: %v", singleDir)
		utils.LogWhenVerbose("Clean: %v", doClean)
		utils.LogWhenVerbose("Print Coverage: %v", doPrint)
		utils.LogWhenVerbose("Ignore Regex: %v", ignorePaths)
		utils.LogWhenVerbose("Slack WebHook: %v", webHook)
		utils.LogWhenVerbose("Prefix: %v", prefix)
		utils.LogWhenVerbose("Depth: %v", depth)
		utils.LogWhenVerbose("Min Coverage: %v", minCoverage)
		utils.LogWhenVerbose("Start Dir: %v", startDir)
		utils.LogWhenVerbose("Path: %v", path)
		utils.LogWhenVerbose("Go Test Args: %v", goTestArgs)
	} else {
		utils.VerboseOff()
	}

	if ignorePaths != "" {
		exclusions = regexp.MustCompile(ignorePaths)
	}

	if depth > 0 && len(prefix) == 0 {
		println("You must specify a prefix when using -depth")
		os.Exit(-1)
	}

	if coverage {
		if singleDir {
			generator.CoverageSingle(path, exclusions, quiet, goTestArgs)
		} else {
			generator.Coverage(path, exclusions, quiet, goTestArgs)
		}
	}

	// switch back to start dir
	err := os.Chdir(startDir)
	if err != nil {
		panic(err)
	}

	coverageOk := true
	if doPrint {
		buffer := bytes.Buffer{}

		if singleDir {
			coverageOk = parser.PrintCoverageSingle(&buffer, path, minCoverage, prefix, depth)
		} else {
			coverageOk = parser.PrintCoverage(&buffer, path, exclusions, minCoverage, prefix, depth)
		}

		fmt.Print(buffer.String())
	}

	if webHook != "" {
		if singleDir {
			parser.SlackCoverageSingle(path, webHook, channelOverride, prefix, depth)
		} else {
			parser.SlackCoverage(path, exclusions, webHook, channelOverride, prefix, depth)
		}
	}

	if doClean {
		cleaner := generator.NewCleaner()
		if singleDir {
			cleaner.Single(path)
		} else {
			cleaner.Recursive(path, exclusions)
		}
	}

	if !coverageOk {
		os.Exit(-1)
	}
}

func getPath() string {
	path := flag.Arg(0)
	if path == "" {
		println("Please include a directory as the last argument")
		os.Exit(-1)
	}
	return path
}

func getGoTestArguments() []string {
	args := flag.Args()

	// We only assume what comes after -- to be `go test` arguments. If there are two arguments, we do not assume them
	// to be `go test` arguments.
	if (len(args) >= 2 && args[1] != "--") || len(args) < 3 {
		return []string{}
	}

	return args[2:]
}
