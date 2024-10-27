package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	api "hugger/apiv2"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	progressbar "github.com/schollz/progressbar/v3"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "help":
		printHelp()
	case "-h":
		printHelp()
	case "--help":
		printHelp()
	case "download":
		handleDownload()
	case "upload":
		handleUpload()
	case "repo-files":
		handleRepoFiles()
	case "meta":
		handleMeta()
	default:
		fmt.Println("Unknown subcommand:", os.Args[1])
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
	fmt.Println("")
	fmt.Println("  repo-files          Perform actions on repository files")
	fmt.Println("    Arguments:")
	fmt.Println("      -repo-id        Repository ID")
	fmt.Println("      -action         Action to perform on repo files")
	fmt.Println("      -file           File to do action with")
	fmt.Println("      -token          A User Access Token generated from https://huggingface.co/settings/tokens")
	fmt.Println("")
	fmt.Println("  meta                Show meta information about repository")
	fmt.Println("    Arguments:")
	fmt.Println("      -repo-id        Repository ID")
	fmt.Println("      -repo-type      Type of repository")
	fmt.Println("      -token          A User Access Token generated from https://huggingface.co/settings/tokens")	
	fmt.Println("")

}

func handleMeta() {
	
	metaf := flag.NewFlagSet("meta", flag.ExitOnError)
	repoID := metaf.String("repo-id", "", "Repository ID")
	repoType := metaf.String("repo-type", "", "Type of the repository")
	token := metaf.String("token", "", "User Access Token generated from https://huggingface.co/settings/tokens")
	metaf.Parse( os.Args[2:] )

	if *repoID == "" || *repoType == "" || *token == "" {
		fmt.Println("meta subcommand requires repo_id, repo-type and token arguments")
		os.Exit(1)
	}

	client := api.HuggingFaceClient{ *token }
	meta, err := client.GetMetadata( *repoType, *repoID )
	if err != nil {
		fmt.Println("Error: failed to get metadata for " + *repoID + " :", err)
	} else {


		description := strings.Replace( strings.Replace( meta.Description, "\n\n", "\n", -1 ), "\n\t\n", "\n", -1 )

		tw := table.NewWriter()
		tw.AppendHeader( table.Row{ meta.Type[3:] + " " + meta.Name + " by " + meta.Creator["name"] } )
		tw.AppendRow( table.Row{ "URL", meta.URL } )
		tw.AppendRow( table.Row{ "License", meta.License } )
		tw.AppendRow( table.Row{ "Description", description } )
		
		kw := ""
		for _, k := range meta.Keywords {
			if strings.Contains( k, "Region" ) {
				k = k[ strings.Index(k, "Region") :]
			}
			kw += k + " "
		}

		tw.AppendFooter( table.Row{ "Keywords", kw } )
		tw.SetStyle( table.StyleColoredDark )
		tw.Style().Color.Header = text.Colors{ text.BgGreen, text.FgBlack }
		tw.Style().Color.Footer = text.Colors{ text.BgGreen, text.FgBlack }

		fmt.Printf("%s\n", tw.Render())
	}
}


func handleDownload() {
	download := flag.NewFlagSet("download", flag.ExitOnError)
	repoID := download.String("repo-id", "", "Repository ID")
	filenames := download.String("filenames", "", "Comma-separated list of filenames")
	repoType := download.String("repo-type", "", "Type of the repository")
	token := download.String("token", "", "User Access Token generated from https://huggingface.co/settings/tokens")
	download.Parse( os.Args[2:] )

	if *repoID == "" || *filenames == "" || *repoType == "" || *token == "" {
		fmt.Println("download subcommand requires repo_id, filenames, and repo-type arguments")
		os.Exit(1)
	}

	client := api.HuggingFaceClient{ *token }
	files := strings.Split( *filenames, "," )
	bar := progressbar.Default( int64(len(files)) )

	for _, file := range files {
		data, err := client.DownloadFile( *repoType, *repoID, file )
		if err != nil {
			fmt.Println("Error downloading file:", err)
			os.Exit(2)
		}else {
			if err := os.WriteFile( file, data, 0666 ); err != nil {
				fmt.Println("Error saving file:", err)
				os.Exit(2)
			}
		}
		bar.Add(1)
	}
}

func handleUpload() {
	upload := flag.NewFlagSet("upload", flag.ExitOnError)
	repoID := upload.String("repo-id", "", "Repository ID")
	filenames := upload.String("filenames", "", "Comma-separated list of filenames")
	repoType := upload.String("repo-type", "", "Type of the repository")
	token := upload.String("token", "", "User Access Token generated from https://huggingface.co/settings/tokens")

	upload.Parse( os.Args[2:] )

	if *repoID == "" || *filenames == "" || *repoType == "" || *token == "" {
		fmt.Println("upload subcommand requires repo_id, filenames, and repo-type arguments")
		os.Exit(1)
	}

	client := api.HuggingFaceClient{ *token }
	files := strings.Split( *filenames, ",")
	bar := progressbar.Default( int64(len(files)) )

	for _, file := range files {
		contents, err := os.ReadFile( file )
		if err != nil {
			fmt.Println("Error reading file", file, ":", err)
			os.Exit(2)
		}


		if err := client.UploadFile( *repoType, *repoID, file, contents ); err != nil {
			fmt.Println("Error downloading file:", err)
			os.Exit(2)
		}
		bar.Add(1)
	}
}

func handleRepoFiles() {

	repoFiles := flag.NewFlagSet("repo-files", flag.ExitOnError)
	repoID := repoFiles.String("repo-id", "", "Repository ID")
	action := repoFiles.String("action", "", "Action to perform on repo files")
	file := repoFiles.String("file", "", "File to do some action with")
	token := repoFiles.String("token", "", "User Access Token generated from https://huggingface.co/settings/tokens")
	repoFiles.Parse( os.Args[2:] )

	if *repoID == "" || *action == "" || *token == "" || *file == "" {
		fmt.Println("repo-files subcommand requires repo_id and action arguments")
		os.Exit(1)
	}

	client := api.HuggingFaceClient{ *token }
	switch *action {
	case "delete":
		if err := client.DeleteFile( "dataset", *repoID, *file ); err != nil {
			fmt.Println("Failed to delete file:", err)
		} else {
			fmt.Println("File successfully deleted")
		}
	default:
		fmt.Println("Error: no such action:", *action)
		fmt.Println("Available actions: {delete}")
	}
}

func Banner() {
	lightyellow := color.New(color.FgCyan, color.Italic).SprintFunc()
	yellow := color.New( color.FgYellow, color.Bold ).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	lightred := color.New(color.FgRed, color.Bold).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()
	black := color.New(color.FgBlack).SprintFunc()

	banner := `                                                                                                                               
                                @@@@@@@@@@@@@@@@@                               
                          @@@@@@@@%###########&@@@@@@@@@                        
                     @@@@@@######%%%%%%%%%%%%%%%######@@@@@@                    
                  @@@@@####%%%%%%%%%%%%%%%%%%%%%%%%%%%####@@@@@                 
                @@@@###%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%###@@@@@              
              @@@###%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%###@@@@            
            @@@###%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%###@@@           
           @@@##%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%##@@@@         
         @@@###%%%%%%%%%%%%.....%%%%%%%%%%%%%%%%.....%%%%%%%%%%%%%###@@@        
        @@@###%%%%%%%%%%%%........%%%%%%%%%%%%........%%%%%%%%%%%%%##%@@@       
        @@@##%%%%%%#####%%....%%.%%%%%%%%%%%%%%.%%....%%%#####%%%%%%##@@@@      
       @@@##%%%%%%%#####%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%####%%%%%%%##@@@      
       @@@##%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%##@@@      
       @@@##%%%%%%%%%%%%%%%%%%,%%%%%%%%%%%%%%%%%%,,%%%%%%%%%%%%%%%%%%##@@@      
       @@@##%%%%%%%%%%%%%%%%%%,,,,,,,%%%%%%%,,,,,,,%%%%%%%%%%%%%%%%%%##@@@      
       @@@##%%%%%%%%%%%%%%%%%%,,,,,,,,,,,,,,,,,,,,%%%%%%%%%%%%%%%%%%%##@@@      
        @@@##%#####%%%%%%%%%%%%%,,,,///,,,///,,,,%%%%%%%%%%%%%%#######@@@@      
     @@@@@@####%%%###%######%%%%%%/////////////%%%%%%######%###%%%%###@@@@@@    
    @@@########%%%%%###%%%%###%%%%%%%%%%%%%%%%%%%%%%##%%%%###%%%%%########@@@@  
    @@###%%%%%####%%%%##%%%%###%%%%%%%%%%%%%%%%%%%###%%%%##%%%%%###%%%%%%##@@@  
   @@@@####%%%%%%###%%%###%%%%###%%%%%%%%%%%%%%%###%%%%###%%%###%%%%%%####@@@@  
   @@###%%%####%%%%%##%%%%%%%%%%###%%%%%%%%%%%###%%%%%%%%%%##%%%%%####%%%%##@@@ 
   @@@###%%%%%%%###%%%%%%%%%%%%%%%##%%%%%%%%%##%%%%%%%%%%%%%%%%##%%%%%%%###@@@@ 
   @@@##%%%####%%%%%%%%%%%%%%%%%%%%##%%%%%%%##%%%%%%%%%%%%%%%%%%%%####%%%###@@@ 
   @@@###%%%%%%%%%%%%%%%%%%%%%%%%%##%%%%%%%%%##%%%%%%%%%%%%%%%%%%%%%%%%%###@@@@ 
     @@@@######%%%%%%%%%%%%%%%%%#################%%%%%%%%%%%%%%%%%######@@@@@   
        @@@@@@@@################%@@@@@@@@@@@@@@@@################@@@@@@@@@      
              @@@@@@@@@@@@@@@@@@@@             @@@@@@@@@@@@@@@@@@@@@            
                                                                                `
	lines := strings.Split( banner, "\n" )
	for _, l := range lines {
		l = strings.Replace(l, "@", white("@"), -1 )
		l = strings.Replace(l, "%", yellow("%"), -1)
		l = strings.Replace(l, "#", lightyellow("#"), -1)
		l = strings.Replace(l, ",", red(","), -1)
		l = strings.Replace(l, "/", lightred("/"), -1)
		l = strings.Replace(l, ".", black("@"), -1)
		fmt.Println(l)
	}
}
