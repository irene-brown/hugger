package apiv2

import (
	"os"
	"fmt"
	"strings"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/k0kubun/go-ansi"
	progressbar "github.com/schollz/progressbar/v3"
)


func ServeRequest( reqType, repoName, repoType, token, action string, files []string, private bool ) error {

	client := HuggingFaceClient{ token }
	
	switch reqType {
	case "meta":
		//fmt.Println("getting metadata...")
		meta, err := client.GetMetadata( repoType, repoName )
		if err != nil {
			return err
		}
		//fmt.Println("collected metadata...")
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
		tw.Style().Color.Header = text.Colors{ text.BgCyan, text.FgWhite, text.Bold }
		tw.Style().Color.Footer = text.Colors{ text.BgCyan, text.FgWhite, text.Bold }

		fmt.Printf("%s\n", tw.Render())

	case "download":
		bar := setupProgressBar( int(len(files)), "Downloading..." )
		for _, file := range files {
			data, err := client.DownloadFile( repoType, repoName, file )
			if err != nil {
				return err
			}else {
				if err := os.WriteFile( file, data, 0666 ); err != nil {
					return err
				}
			}
			bar.Add(1)
		}
		fmt.Println()
	case "upload":
		bar := setupProgressBar( int(len(files)), "Uploading..." )

		for _, file := range files {
			contents, err := os.ReadFile( file )
			if err != nil {
				return err
			}

			if err := client.UploadFile( repoType, repoName, file, contents ); err != nil {
				return err
			}
			bar.Add(1)
		}
		fmt.Println()

	case "repo":
		switch action {
		case "create":
			if err := client.CreateRepo( repoType, repoName, private ); err != nil {
				return err
			}
			fmt.Println("Repo was created")
		case "delete":
			if err := client.DeleteRepo( repoName ); err != nil {
				return err
			}
			fmt.Println("Repo was deleted")
		}
	case "repo-files":
		switch action {
		case "delete":
			if len(files) < 1 {
				return nil
			}
			if err := client.DeleteFile( repoType, repoName, files[0] ); err != nil {
				return err
			}
		case "list":
			file := "/"
			if len(files) > 0 {
				if files[0] != "" {
					file = files[0]
				}
			}
			repoFiles, err := client.ListFilesInRepo( repoType, repoName, file, false )
			if err != nil {
				return err
			}
			
			tw := table.NewWriter()
			tw.AppendHeader( table.Row{"Directory listing of " + file} )
			for _, file := range repoFiles {
				tw.AppendRow( table.Row{ file } )
			}
			tw.AppendFooter( table.Row{"Total", fmt.Sprintf("%d", len(repoFiles))} )
			tw.SetStyle( table.StyleColoredDark )
			tw.Style().Color.Header = text.Colors{ text.BgCyan, text.FgWhite, text.Bold }
			tw.Style().Color.Footer = text.Colors{ text.BgCyan, text.FgWhite, text.Bold }

			fmt.Printf("%s\n", tw.Render())
		default:
			return fmt.Errorf("Error: no such action: %s. Choose from {delete,list}\n")
		}
	default:
		return fmt.Errorf("Invalid command: %s\n", reqType)
	}
	return nil
}

func setupProgressBar( x int, msg string ) *progressbar.ProgressBar {
	bar := progressbar.NewOptions( x,
		progressbar.OptionSetWriter( ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetDescription(
			fmt.Sprintf("[cyan] %s [reset]", msg ),
		),
		progressbar.OptionSetTheme( progressbar.Theme{
			Saucer:	"[green]=[reset]",
			SaucerHead: "[yellow]>[reset]",
			SaucerPadding: " ",
			BarStart: "[",
			BarEnd: "]",
		} ),
	)
	return bar
}
