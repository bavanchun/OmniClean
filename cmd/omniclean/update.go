package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

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

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return fmt.Errorf("failed to create updater: %w", err)
	}

	latest, found, err := updater.DetectLatest(context.Background(), selfupdate.ParseSlug("bavanchun/OmniClean"))
	if err != nil {
		return fmt.Errorf("failed to check latest version: %w", err)
	}
	if !found {
		fmt.Println("No releases found.")
		return nil
	}

	if latest.LessOrEqual(current) {
		fmt.Printf("Already up to date (%s).\n", current)
		return nil
	}

	fmt.Printf("Updating to %s...\n", latest.Version())
	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return fmt.Errorf("could not locate executable: %w", err)
	}

	if err := updater.UpdateTo(context.Background(), latest, exe); err != nil {
		if isPermissionError(err) {
			return reExecWithSudo()
		}
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Printf("Successfully updated to %s.\n", latest.Version())
	return nil
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
