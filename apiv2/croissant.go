package apiv2

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/k0kubun/go-ansi"
	progressbar "github.com/schollz/progressbar/v3"
)

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

func ServeRequest(reqType, repoName, repoType, token, action, split string, files []string, private bool) error {
	client := HuggingFaceClient{Token: token}

	switch reqType {
	case "meta":
		meta, err := client.GetMetadata(repoType, repoName)
		if err != nil {
			return err
		}
		displayMetadata(meta)

	case "statistics":
		stat, err := client.GetDatasetStatistics( repoName, split )
		if err != nil {
			return fmt.Errorf("failed to get statistics for %s: %s", repoName, err)
		}
		displayStatistics(stat, repoName)

	case "download":
		if err := processFiles(client, files, repoType, repoName, "download"); err != nil {
			return err
		}

	case "upload":
		if err := processFiles(client, files, repoType, repoName, "upload"); err != nil {
			return err
		}

	case "repo":
		if err := manageRepo(client, repoType, repoName, action, private); err != nil {
			return err
		}

	case "repo-files":
		if err := manageRepoFiles(client, repoType, repoName, files, action); err != nil {
			return err
		}

	default:
		return fmt.Errorf("invalid command: %s", reqType)
	}
	return nil
}

func displayMetadata(meta *MetadataResponse) {

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
	tw.Style().Color.Header = text.Colors{ text.BgBlue, text.FgWhite, text.Bold }
	tw.Style().Color.Footer = text.Colors{ text.BgBlue, text.FgWhite, text.Bold }

	fmt.Println(tw.Render())
}

func displayStatistics(stat *Statistics, repoName  string) {

	tw := table.NewWriter()
	tw.AppendHeader(table.Row{ fmt.Sprintf("Statistics for dataset %s", repoName) })
	tw.AppendRow( table.Row{"Partial", fmt.Sprintf("%s", stat.Partial)} )
	tw.AppendRow( table.Row{"NumExamples", fmt.Sprintf("%s", stat.NumExamples)} )

	for k, v := range stat.Statistics {
		tw.AppendRow( table.Row{ k, fmt.Sprintf("%s", v) } )
	}
	
	tw.SetStyle(table.StyleColoredDark)
	tw.Style().Color.Header = text.Colors{ text.BgCyan, text.FgWhite, text.Bold }
	tw.Style().Color.Footer = text.Colors{ text.BgCyan, text.FgWhite, text.Bold }
	fmt.Println(tw.Render())
}

/*
 * getGradientColor - show progress not only using progressbar
 * but also using colorful text animations
 */
func getGradientColor( progress float64 ) string {
	r := int( 255 * (1 - progress) )
	g := int( 255 * progress )
	return fmt.Sprintf("\033[38;2;%d;%d;0m", r, g)
}


func processFiles(client HuggingFaceClient, files []string, repoType, repoName, action string) error {

	bar := progressbar.NewOptions(len(files),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription(fmt.Sprintf("[red]Processing %s...[reset]", action)),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[blue]-[reset]",
			SaucerHead:    "[cyan][bold]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	totalSteps := len(files)

	for i, file := range files {

		color := getGradientColor( float64(i+1) / float64(totalSteps) )

		switch action {
		case "download":
			content, err := client.DownloadFile(repoType, repoName, file)
			if err != nil {
				return fmt.Errorf("failed to download %s: %v", file, err)
			}
			if err := ioutil.WriteFile(file, content, 0644); err != nil {
				return fmt.Errorf("failed to save %s: %v", file, err)
			}

		case "upload":
			content, err := ioutil.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read %s: %v", file, err)
			}
			if err := client.UploadFile(repoType, repoName, file, content); err != nil {
				return fmt.Errorf("failed to upload %s: %v", file, err)
			}
		}
		bar.Describe( fmt.Sprintf("%s Processing %s...[reset]", color, action) )
		bar.Add(1)
	}
	return nil

}

func manageRepo(client HuggingFaceClient, repoType, repoName, action string, private bool) error {
	switch action {
	case "create":
		if err := client.CreateRepo(repoType, repoName, private); err != nil {
			return fmt.Errorf("failed to create repository: %v", err)
		}
		fmt.Printf("✨ Repository %s/%s created successfully!\n", repoType, repoName)

	case "delete":
		if err := client.DeleteRepo(repoName); err != nil {
			return fmt.Errorf("failed to delete repository: %v", err)
		}
		fmt.Printf("🗑️  Repository %s/%s deleted successfully!\n", repoType, repoName)

	default:
		return fmt.Errorf("invalid repo action: %s", action)
	}
	return nil
}

func manageRepoFiles(client HuggingFaceClient, repoType, repoName string, files []string, action string) error {
	switch action {
	case "list":
		filepath := "/"
		if len(files) > 0 {
			if len(files[0]) > 0 {
				filepath = files[0]
			}
		}

		repoFiles, err := client.ListFilesInRepo(repoType, repoName, filepath, false)
		if err != nil {
			return fmt.Errorf("failed to list files: %v", err)
		}

		tw := table.NewWriter()
		tw.AppendHeader( table.Row{"Directory listing of " + filepath} )

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

		//tw.AppendFooter( table.Row{"Folders:", fmt.Sprintf("%d", directoriesCount) } )
		//tw.AppendFooter( table.Row{"Files:", fmt.Sprintf("%d", filesCount) } )
		tw.AppendFooter( table.Row{"Total", fmt.Sprintf("%d", len(repoFiles))} )

		tw.SetStyle( table.StyleColoredDark )
		tw.Style().Color.Header = text.Colors{ text.BgBlue, text.FgWhite, text.Bold }
		tw.Style().Color.Footer = text.Colors{ text.BgBlue, text.FgWhite, text.Bold }

		fmt.Println(tw.Render())

	case "delete":
		for _, file := range files {
			if err := client.DeleteFile(repoType, repoName, file); err != nil {
				return fmt.Errorf("failed to delete %s: %v", file, err)
			}
			fmt.Printf("🗑️  Deleted %s\n", file)
		}

	default:
		return fmt.Errorf("invalid file action: %s", action)
	}
	return nil
}
