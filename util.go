package main

import (
	_ "embed"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
)

var (
	//go:embed "img/share-android.svg"
	shareAndroidBytes []byte

	//go:embed "img/share.svg"
	shareBytes []byte
)

func durationSince(fromString string) string {
	from, err := time.Parse(time.RFC3339, fromString)
	if err != nil {
		fyne.LogError("Failed to parse time", err)
		return "now"
	}

	gap := time.Since(from)
	switch {
	case gap < time.Minute:
		return "now"
	case gap < time.Hour:
		return fmt.Sprintf("%.0f minutes ago", gap.Minutes())
	case gap < time.Hour*24:
		return fmt.Sprintf("%.0f hours ago", gap.Minutes()/60)
	default:
		return fmt.Sprintf("%.0f days ago", gap.Minutes()/60/24)
	}
}
