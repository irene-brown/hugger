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

// parse description, drop all unneeded characters
// also add a little bit of markup
func prepareDescription( description string ) string {

	description = strings.Replace( strings.Replace( description, "\n\n", "\n", -1 ), "\n\t", "\n", -1 )
	description = strings.Replace( description, "\t", "", -1 )
	description = strings.Replace( description, "  ", " ", -1 )
	description = strings.Replace( description, "\r\n", "", -1 )
	description = strings.Replace( description, "\r", "", -1 )
	description = strings.Trim( description, " " )
	description = strings.Trim( description, "\n" )
	description = strings.Trim( description, "\t" )
	for {
		if strings.Contains( description, "\n\n\n" ) {
			description = strings.Replace( description, "\n\n\n", "\n\n", -1 )
		} else {
			break
		}
	}
	
	res := ""
	tmpRes := strings.Split( description, "\n\n" )
	for i, r := range tmpRes {
		if i % 2 == 0 {
			res += "\033[1;1m" + "### " + r + "\n"
		} else {
			tmp := strings.Split(r, "\n")
			for j, _ := range tmp {
				tmp[j] = "\t" + tmp[j]
			}
			r = strings.Join( tmp, "\n" )
			res += r + "\n\n"
		}
	}
	res = strings.Trim( res, "\n" )

	return res
}

func ServeRequest( reqType, repoName, repoType, token, action string, files []string, private bool ) error {

	client := HuggingFaceClient{ token }
	
	switch reqType {
	case "meta":
		meta, err := client.GetMetadata( repoType, repoName )
		if err != nil {
			return err
		}

		description := prepareDescription( meta.Description )


		tw := table.NewWriter()

		reset := "\x1b[39m"

		tw.AppendHeader( table.Row{ meta.Type[3:] + " " + meta.Name + " by " + meta.Creator["name"] } )
		tw.AppendRow( table.Row{ "Name", "\033[38;2;150;200;200;1m" + meta.Name + reset } )
		tw.AppendRow( table.Row{ "Type", "\033[38;2;100;200;200;1m" + meta.Type[3:] + reset } )
		tw.AppendRow( table.Row{ "Author","\033[38;2;50;200;200;1m" + meta.Creator["name"] + reset } )
		tw.AppendRow( table.Row{ "URL", "\033[38;2;0;200;200;1m" + meta.URL + reset} )
		tw.AppendRow( table.Row{ "License", "\033[38;2;0;150;200;1m" + meta.License + reset } )
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
		totalSteps := len(files)

		for i, file := range files {
			
			color := getGradientColor( float64(i+1) / float64(totalSteps) )

			data, err := client.DownloadFile( repoType, repoName, file )
			if err != nil {
				return err
			}else {
				if err := os.WriteFile( file, data, 0666 ); err != nil {
					return err
				}
			}

			bar.Describe( fmt.Sprintf("%s Downloading... [reset]", color) )
			bar.Add(1)
		}
		fmt.Println()

	case "upload":
		bar := setupProgressBar( int(len(files)), "Uploading..." )
		totalSteps := len(files)

		for i, file := range files {
			
			color := getGradientColor( float64(i+1) / float64(totalSteps) )
			
			contents, err := os.ReadFile( file )
			if err != nil {
				return err
			}

			if err := client.UploadFile( repoType, repoName, file, contents ); err != nil {
				return err
			}
			
			bar.Describe( fmt.Sprintf("%s Uploading... [reset]", color) )
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

			directoriesCount := 0
			filesCount := 0

			for _, file := range repoFiles {
				if file[len(file)-1] == '/' { // directory
					file = "\033[38;2;0;200;200;1m" + file 
					directoriesCount++
				} else {
					filesCount++
				}
				tw.AppendRow( table.Row{ file } )
			}
			tw.AppendFooter( table.Row{"Folders:", fmt.Sprintf("%d", directoriesCount) } )
			tw.AppendFooter( table.Row{"Files:", fmt.Sprintf("%d", filesCount) } )
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
			Saucer:	"[green]-[reset]",
			SaucerHead: "[yellow][bold]>[reset]",
			SaucerPadding: " ",
			BarStart: "[",
			BarEnd: "]",
		} ),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth( 80 ),
	)
	return bar
}

// return gradient for progressbar
func getGradientColor( progress float64 ) string {
	r := int( 255 * (1 - progress) )
	g := int( 255 * progress )
	return fmt.Sprintf("\033[38;2;%d;%d;0m", r, g)
}
