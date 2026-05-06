package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update omniclean to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(version)
		},
	}
}

func runUpdate(current string) error {
	fmt.Printf("Current version: %s\n", current)
	fmt.Println("Checking for updates...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return fmt.Errorf("failed to create updater: %w", err)
	}

	latest, found, err := updater.DetectLatest(ctx, selfupdate.ParseSlug("bavanchun/OmniClean"))
	if err != nil {
		return fmt.Errorf("failed to check latest version: %w", err)
	}
	if !found {
		fmt.Println("No releases found.")
		return nil
	}

	needsUpdate, err := isUpdateAvailable(current, latest.Version())
	if err != nil {
		return err
	}
	if !needsUpdate {
		fmt.Printf("Already up to date (%s).\n", current)
		return nil
	}

	fmt.Printf("Updating to %s...\n", latest.Version())
	fmt.Println("Downloading and replacing binary; this can take a minute on slow connections.")
	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return fmt.Errorf("could not locate executable: %w", err)
	}

	if err := updater.UpdateTo(ctx, latest, exe); err != nil {
		if isPermissionError(err) {
			return reExecWithSudo()
		}
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Printf("Successfully updated to %s.\n", latest.Version())
	return nil
}

func isUpdateAvailable(current, latest string) (bool, error) {
	latestVersion, err := semver.NewVersion(latest)
	if err != nil {
		return false, fmt.Errorf("latest release has invalid version %q: %w", latest, err)
	}

	currentVersion, err := semver.NewVersion(current)
	if err != nil {
		return true, nil
	}

	return currentVersion.LessThan(latestVersion), nil
}

func isPermissionError(err error) bool {
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return errors.Is(pathErr.Err, syscall.EACCES)
	}
	return false
}

func reExecWithSudo() error {
	fmt.Println("Permission denied — re-running with sudo...")
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate executable: %w", err)
	}
	cmd := exec.Command("sudo", exe, "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
