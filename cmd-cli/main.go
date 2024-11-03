package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	api "hugger/apiv2"
	huggerLog "hugger/log"
	"github.com/fatih/color"
)

type HError struct {
	Error string `json:"error"`
}

func main() {
	// Check for updates
	api.UpdateApp()

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	// Handle subcommands
	switch os.Args[1] {
	case "help", "-h", "--help":
		printHelp()
	case "download":
		handleDownload()
	case "upload":
		handleUpload()
	case "repo":
		handleRepo()
	case "repo-files":
		handleRepoFiles()
	case "meta":
		handleMeta()
	case "statistics":
		handleStatistics()
	default:
		fmt.Printf("Unknown subcommand: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func printHelp() {
	Banner()
	fmt.Println("Available subcommands:")
	fmt.Println("  help                Show this help message")
	fmt.Println("  download            Download files from a repository")
	fmt.Println("    Arguments:")
	fmt.Println("      -repo-id        Repository ID")
	fmt.Println("      -filenames      Comma-separated list of filenames")
	fmt.Println("      -repo-type      Type of the repository")
	fmt.Println("      -token          A User Access Token generated from https://huggingface.co/settings/tokens")
	fmt.Println()
	fmt.Println("  upload              Upload files to a repository")
	fmt.Println("    Arguments:")
	fmt.Println("      -repo-id        Repository ID")
	fmt.Println("      -filenames      Comma-separated list of filenames")
	fmt.Println("      -repo-type      Type of the repository")
	fmt.Println("      -token          A User Access Token generated from https://huggingface.co/settings/tokens")
	fmt.Println()
	fmt.Println("  repo                Perform actions on repository")
	fmt.Println("    Arguments:")
	fmt.Println("      -repo-id        Repository ID")
	fmt.Println("      -repo-type      Type of the repository")
	fmt.Println("      -action         Action to perform on repo files ({delete,create})")
	fmt.Println("      -token          A User Access Token generated from https://huggingface.co/settings/tokens")
	fmt.Println("      -private        Create/delete private repository")
	fmt.Println()
	fmt.Println("  repo-files          Perform actions on repository files")
	fmt.Println("    Arguments:")
	fmt.Println("      -repo-id        Repository ID")
	fmt.Println("      -repo-type      Type of the repository")
	fmt.Println("      -action         Action to perform on repo files ({delete,list})")
	fmt.Println("      -file           File to do action with. Optionally, you can pass a directory name here")
	fmt.Println("      -token          A User Access Token generated from https://huggingface.co/settings/tokens")
	fmt.Println()
	fmt.Println("  meta                Show meta information about repository")
	fmt.Println("    Arguments:")
	fmt.Println("      -repo-id        Repository ID")
	fmt.Println("      -repo-type      Type of repository")
	fmt.Println("      -token          A User Access Token generated from https://huggingface.co/settings/tokens")
	fmt.Println()
	fmt.Println("  statistics          Show statistics for specified repository. Dataset-only feature")
	fmt.Println("    Arguments:")
	fmt.Println("      -repo-id        Repository ID")
	fmt.Println("      -split          Dataset split (e.g. train)")
	fmt.Println("      -token          A User Access Token generated from https://huggingface.co/settings/tokens")
	fmt.Println()
}

func handleMeta() {
	metaf := flag.NewFlagSet("meta", flag.ExitOnError)
	repoID := metaf.String("repo-id", "", "Repository ID")
	repoType := metaf.String("repo-type", "", "Type of the repository")
	token := metaf.String("token", "", "User Access Token")

	metaf.Parse(os.Args[2:])

	if *repoID == "" || *repoType == "" || *token == "" {
		fmt.Println("meta subcommand requires repo-id, repo-type, and token arguments")
		os.Exit(1)
	}

	if err := api.ServeRequest("meta", *repoID, *repoType, *token, "", "", nil, false); err != nil {
		handleError(err)
	}
}

func handleStatistics() {
	stat := flag.NewFlagSet("statistics", flag.ExitOnError)
	repoID := stat.String("repo-id", "", "Repository ID")
	split := stat.String("split", "", "Dataset split(e.g. train)")
	token := stat.String("token", "", "User Access Token")

	stat.Parse( os.Args[2:] )
	if *repoID == "" || *token == "" || *split == "" {
		fmt.Println("statistics subcommand requires repo-id, split and token arguments")
		os.Exit(1)
	}
	if err := api.ServeRequest( "statistics", *repoID, "dataset", *token, "", *split, nil, false ); err != nil {
		handleError(err)
	}
}

func handleDownload() {
	download := flag.NewFlagSet("download", flag.ExitOnError)
	repoID := download.String("repo-id", "", "Repository ID")
	filenames := download.String("filenames", "", "Comma-separated list of filenames")
	repoType := download.String("repo-type", "", "Type of the repository")
	token := download.String("token", "", "User Access Token")

	download.Parse(os.Args[2:])

	if *repoID == "" || *filenames == "" || *repoType == "" || *token == "" {
		fmt.Println("download subcommand requires repo-id, filenames, repo-type, and token arguments")
		os.Exit(1)
	}

	files := strings.Split(*filenames, ",")
	if err := api.ServeRequest("download", *repoID, *repoType, *token, "", "", files, false); err != nil {
		handleError(err)
	}
}

func handleUpload() {
	upload := flag.NewFlagSet("upload", flag.ExitOnError)
	repoID := upload.String("repo-id", "", "Repository ID")
	filenames := upload.String("filenames", "", "Comma-separated list of filenames")
	repoType := upload.String("repo-type", "", "Type of the repository")
	token := upload.String("token", "", "User Access Token")

	upload.Parse(os.Args[2:])

	if *repoID == "" || *filenames == "" || *repoType == "" || *token == "" {
		fmt.Println("upload subcommand requires repo-id, filenames, repo-type, and token arguments")
		os.Exit(1)
	}

	files := retrieveFiles(*filenames)
	if err := api.ServeRequest("upload", *repoID, *repoType, *token, "", "", files, false); err != nil {
		handleError(err)
	}
}

func handleRepo() {
	repo := flag.NewFlagSet("repo", flag.ExitOnError)
	repoID := repo.String("repo-id", "", "Repository ID")
	repoType := repo.String("repo-type", "", "Type of the repository")
	action := repo.String("action", "", "Action to perform on repo files")
	token := repo.String("token", "", "User Access Token")
	private := repo.Bool("private", false, "Flag for private repositories")

	repo.Parse(os.Args[2:])

	if *repoID == "" || *repoType == "" || *action == "" || *token == "" {
		fmt.Println("repo subcommand requires repo-id, repo-type, action, and token arguments")
		os.Exit(1)
	}

	if err := api.ServeRequest("repo", *repoID, *repoType, *token, *action, "", nil, *private); err != nil {
		handleError(err)
	}
}

func handleRepoFiles() {
	repoFiles := flag.NewFlagSet("repo-files", flag.ExitOnError)
	repoID := repoFiles.String("repo-id", "", "Repository ID")
	repoType := repoFiles.String("repo-type", "", "Type of the repository")
	action := repoFiles.String("action", "", "Action to perform on repo files")
	file := repoFiles.String("file", "", "File to do some action with. Optionally, you can pass a directory here")
	token := repoFiles.String("token", "", "User Access Token")

	repoFiles.Parse(os.Args[2:])

	if *repoID == "" || *repoType == "" || *action == "" || *token == "" {
		fmt.Println("repo-files subcommand requires repo-id, repo-type, token, and action arguments")
		os.Exit(1)
	}

	files := retrieveFiles(*file)
	if err := api.ServeRequest("repo-files", *repoID, *repoType, *token, *action, "", files, false); err != nil {
		handleError(err)
	}
}

func retrieveFiles(filenames string) []string {
	res := []string{}
	tmpres := strings.Split(filenames, ",")

	for _, fn := range tmpres {
		fileInfo, err := os.Stat(fn)
		if err == nil {
			if fileInfo.IsDir() {
				dirFiles := listAllFiles(fn)
				res = append(res, dirFiles...)
			} else {
				res = append(res, fn)
			}
		} else {
			res = append(res, fn) // Add the filename if it does not exist
		}
	}

	if len(res) == 0 {
		return tmpres // Return original list if no files found
	}
	return res
}

func listAllFiles(folder string) []string {
	var res []string

	err := filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path != folder {
			res = append(res, path)
		}
		return nil
	})
	if err != nil {
		return nil // Return nil on error
	}
	return res
}

func handleError(err error) {
	var e HError
	if nerr := json.Unmarshal([]byte(err.Error()), &e); nerr != nil {
		huggerLog.Error( err.Error() ) //fmt.Println("Error:", err)
	} else {
		huggerLog.Error(e.Error)
	}
}

func isTerminal() bool {
	switch runtime.GOOS {
	case "windows":
		return os.Getenv("TERM") == "xterm" || os.Getenv("ANSICON") != ""
	case "darwin", "linux":
		return true
	default:
		return false
	}
}

func Banner() {

	lines := []string{"",
		"                                @@@@@@@@@@@@@@@@@                               ",
		"                          @@@@@@@@#############@@@@@@@@@                        ",
		"                     @@@@@@######%%%%%%%%%%%%%%%######@@@@@@                    ",
		"                  @@@@@####%%%%%%%%%%%%%%%%%%%%%%%%%%%####@@@@@                 ",
		"                @@@@###%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%###@@@@@              ",
		"              @@@###%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%###@@@@            ",
		"            @@@###%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%###@@@           ",
		"           @@@##%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%##@@@@         ",
		"         @@@###%%%%%%%%%%%%.....%%%%%%%%%%%%%%%%.....%%%%%%%%%%%%%###@@@        ",
		"        @@@###%%%%%%%%%%%%........%%%%%%%%%%%%........%%%%%%%%%%%%%###@@@       ",
		"        @@@##%%%%%%#####%%....%%.%%%%%%%%%%%%%%.%%....%%%#####%%%%%%##@@@@      ",
		"       @@@##%%%%%%%#####%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%####%%%%%%%##@@@      ",
		"       @@@##%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%##@@@      ",
		"       @@@##%%%%%%%%%%%%%%%%%%,%%%%%%%%%%%%%%%%%%,,%%%%%%%%%%%%%%%%%%##@@@      ",
		"       @@@##%%%%%%%%%%%%%%%%%%,,,,,,,%%%%%%%,,,,,,,%%%%%%%%%%%%%%%%%%##@@@      ",
		"       @@@##%%%%%%%%%%%%%%%%%%,,,,,,,,,,,,,,,,,,,,%%%%%%%%%%%%%%%%%%%##@@@      ",
		"        @@@##%#####%%%%%%%%%%%%%,,,,///,,,///,,,,%%%%%%%%%%%%%%#######@@@@      ",
		"     @@@@@@####%%%###%######%%%%%%/////////////%%%%%%######%###%%%%###@@@@@@    ",
		"    @@@########%%%%%###%%%%###%%%%%%%%%%%%%%%%%%%%%%##%%%%###%%%%%########@@@@  ",
		"    @@###%%%%%####%%%%##%%%%###%%%%%%%%%%%%%%%%%%%###%%%%##%%%%%###%%%%%%##@@@  ",
		"   @@@@####%%%%%%###%%%###%%%%###%%%%%%%%%%%%%%%###%%%%###%%%###%%%%%%####@@@@  ",
		"   @@###%%%####%%%%%##%%%%%%%%%%###%%%%%%%%%%%###%%%%%%%%%%##%%%%%####%%%%##@@@ ",
		"   @@@###%%%%%%%###%%%%%%%%%%%%%%%##%%%%%%%%%##%%%%%%%%%%%%%%%%##%%%%%%%###@@@@ ",
		"   @@@##%%%####%%%%%%%%%%%%%%%%%%%%##%%%%%%%##%%%%%%%%%%%%%%%%%%%%####%%%###@@@ ",
		"   @@@###%%%%%%%%%%%%%%%%%%%%%%%%%##%%%%%%%%%##%%%%%%%%%%%%%%%%%%%%%%%%%###@@@@ ",
		"     @@@@######%%%%%%%%%%%%%%%%%#################%%%%%%%%%%%%%%%%%######@@@@@   ",
		"        @@@@@@@@#################@@@@@@@@@@@@@@@@################@@@@@@@@@      ",
		"              @@@@@@@@@@@@@@@@@@@@             @@@@@@@@@@@@@@@@@@@@@            ",
		"                                                                                "}
	for _, l := range lines {
		if isTerminal() { // if terminal supports ANSI colors, make banner prettier

			lightyellow := color.New(color.FgCyan, color.Italic).SprintFunc()
			yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()
			lightred := color.New(color.FgRed, color.Bold).SprintFunc()
			white := color.New(color.FgWhite).SprintFunc()
			black := color.New(color.FgBlack).SprintFunc()

			l = strings.Replace(l, "@", white("@"), -1)
			l = strings.Replace(l, "%", yellow("%"), -1)
			l = strings.Replace(l, "#", lightyellow("#"), -1)
			l = strings.Replace(l, ",", red(","), -1)
			l = strings.Replace(l, "/", lightred("/"), -1)
			l = strings.Replace(l, ".", black("@"), -1)
		}
		fmt.Println(l)
	}
}
