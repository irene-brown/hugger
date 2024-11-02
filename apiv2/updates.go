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
	appName        = "hugger"
	currentVersion = "0.2.0" // Current version of Hugger
	// URL to check for the latest version
	updateURL = "https://raw.githubusercontent.com/irene-brown/hugger/refs/heads/main/VERSION"
)

// checkForUpdate checks for the latest version of the application.
func checkForUpdate() (string, error) {
	resp, err := http.Get(updateURL)
	if err != nil {
		return "", fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	var latestVersion string
	_, err = fmt.Fscan(resp.Body, &latestVersion)
	if err != nil {
		return "", fmt.Errorf("failed to read latest version: %w", err)
	}

	return latestVersion, nil
}

// downloadUpdate downloads the specified version of the application.
func downloadUpdate(version string) error {
	downloadURL := fmt.Sprintf("%s/%s/releases/download/%s/%s_%s", updatesBaseURL, appName, version, appName, runtime.GOOS)
	if runtime.GOOS == "windows" {
		// Windows-specific executable extension
		downloadURL += ".exe"
	}

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	thisBinary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	out, err := os.Create(thisBinary)
	if err != nil {
		return fmt.Errorf("failed to create binary file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write update to file: %w", err)
	}

	return os.Chmod(thisBinary, 0755) // Make it executable
}

// UpdateApp checks for updates and applies them if a new version is available.
func UpdateApp() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
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
		if err := exec.Command(executable).Start(); err != nil {
			return fmt.Errorf("failed to restart application: %w", err)
		}
		os.Exit(0)
	}
	return nil
}
