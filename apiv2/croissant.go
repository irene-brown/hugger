package apiv2

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/k0kubun/go-ansi"
	progressbar "github.com/schollz/progressbar/v3"
)

func prepareDescription(description string) string {
	description = strings.ReplaceAll(description, "\n\n", "\n")
	description = strings.ReplaceAll(description, "\t", "")
	description = strings.ReplaceAll(description, "  ", " ")
	description = strings.ReplaceAll(description, "\r\n", "")
	description = strings.ReplaceAll(description, "\r", "")
	description = strings.TrimSpace(description)

	var result string
	sections := strings.Split(description, "\n\n")
	for i, section := range sections {
		if i%2 == 0 {
			result += "\033[1;1m### " + section + "\n"
		} else {
			lines := strings.Split(section, "\n")
			for j := range lines {
				lines[j] = "\t" + lines[j]
			}
			result += strings.Join(lines, "\n") + "\n\n"
		}
	}
	return strings.Trim(result, "\n")
}

func ServeRequest(reqType, repoName, repoType, token, action string, files []string, private bool) error {
	client := HuggingFaceClient{Token: token}

	switch reqType {
	case "meta":
		meta, err := client.GetMetadata(repoType, repoName)
		if err != nil {
			return err
		}
		displayMetadata(meta)

	case "download":
		if err := processFiles(client, files, repoType, repoName, "download"); err != nil {
			return err
		}

	case "upload":
		if err := processFiles(client, files, repoType, repoName, "upload"); err != nil {
			return err
		}

	case "repo":
		if err := manageRepo(client, repoType, repoName, action); err != nil {
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
	description := prepareDescription(meta.Description)
	tw := table.NewWriter()

	tw.AppendHeader(table.Row{meta.Type[3:] + " " + meta.Name + " by " + meta.Creator["name"]})
	tw.AppendRow(table.Row{"Name", "\033[38;2;150;200;200;1m" + meta.Name + "\x1b[39m"})
	tw.AppendRow(table.Row{"Type", "\033[38;2;100;200;200;1m" + meta.Type[3:] + "\x1b[39m"})
	tw.AppendRow(table.Row{"Author", "\033[38;2;50;200;200;1m" + meta.Creator["name"] + "\x1b[39m"})
	tw.AppendRow(table.Row{"License", "\033[38;2;0;200;200;1m" + meta.License + "\x1b[39m"})
	tw.AppendRow(table.Row{"Description", description})

	if len(meta.Keywords) > 0 {
		tw.AppendRow(table.Row{"Keywords", "\033[38;2;200;200;0;1m" + strings.Join(meta.Keywords, ", ") + "\x1b[39m"})
	}

	tw.SetStyle(table.StyleRounded)
	fmt.Println(tw.Render())
}

func processFiles(client HuggingFaceClient, files []string, repoType, repoName, action string) error {
	bar := progressbar.NewOptions(len(files),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription(fmt.Sprintf("[cyan]Processing %s...[reset]", action)),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	for _, file := range files {
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
		bar.Add(1)
	}
	return nil
}

func manageRepo(client HuggingFaceClient, repoType, repoName, action string) error {
	switch action {
	case "create":
		if err := client.CreateRepo(repoType, repoName); err != nil {
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
		fileList, err := client.ListFilesInRepo(repoType, repoName, "", true)
		if err != nil {
			return fmt.Errorf("failed to list files: %v", err)
		}

		tw := table.NewWriter()
		tw.AppendHeader(table.Row{"File Path"})
		for _, file := range fileList {
			tw.AppendRow(table.Row{file})
		}
		tw.SetStyle(table.StyleRounded)
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
