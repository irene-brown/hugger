package apiv2

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

const (
	updatesBaseURL = "https://github.com/irene-brown/"
	appName = "hugger"
	currentVersion = "0.2.0" // Current version of Hugger
	// URL to check for the latest version
	updateURL      = "https://raw.githubusercontent.com/irene-brown/hugger/refs/heads/main/VERSION"
)

func checkForUpdate() (string, error) {
	resp, err := http.Get(updateURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var latestVersion string
	_, err = fmt.Fscan(resp.Body, &latestVersion)
	if err != nil {
		return "", err
	}

	return latestVersion, nil
}

func downloadUpdate(version string) error {
	downloadURL := fmt.Sprintf("%s/%s/releases/download/%s/%s_%s", updatesBaseURL, appName, version, appName, runtime.GOOS)
	if runtime.GOOS == "windows" {
		// windows-only extension
		downloadURL += ".exe"
	}
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	thisBinary, err := os.Executable()
	if err != nil {
		return err
	}

	out, err := os.Create( thisBinary )
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return os.Chmod( thisBinary, 0755 ) // Make it executable
}

/*
 * Update hugger if new version is already available
 */
func UpdateApp() error {

	executable, err := os.Executable()
	if err != nil {
		return err
	}
	latestVersion, err := checkForUpdate()
	if err != nil {
		fmt.Println("Error checking for updates:", err)
		return err
	}

	if latestVersion != currentVersion {
		fmt.Printf("New version available: %s (current: %s)\n", latestVersion, currentVersion)
		err = downloadUpdate(latestVersion)
		if err != nil {
			fmt.Println("Error downloading update:", err)
			return err
		}
		fmt.Println("Update downloaded. Restarting application...")

		// Restart the application
		exec.Command( executable ).Start()
		os.Exit(0)
	}
	return nil
}
