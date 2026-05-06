//go:build !darwin

package main

import "context"

func runUninstallApps(_ context.Context, _ bool) error { return nil }
