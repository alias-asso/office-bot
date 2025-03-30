package main

import (
	"errors"
	"io"
	"os"
	"strings"
)

var statusFile *os.File

func init() {
	f, err := os.OpenFile(config.StatusFilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	statusFile = f
}

func GetStatus() (bool, error) {
	// Reset read position to start
	_, err := statusFile.Seek(0, 0)
	if err != nil {
		return false, err
	}

	status, err := io.ReadAll(statusFile)
	if err != nil {
		return false, err
	}

	switch strings.TrimSpace(string(status)) {
	case "open":
		return true, nil
	case "closed":
		return false, nil
	default:
		return false, errors.New("Statut inconnu.")
	}
}

func SetStatus(status bool) error {
	err := statusFile.Truncate(0)
	if err != nil {
		return err
	}

	_, err = statusFile.Seek(0, 0)
	if err != nil {
		return err
	}

	var content string
	if status {
		content = "open"
	} else {
		content = "closed"
	}

	_, err = io.WriteString(statusFile, content)
	if err != nil {
		return err
	}

	return statusFile.Sync()
}

func ToggleStatus() (string, error) {

	status, err := GetStatus()
	if err != nil {
		return "", errors.New("Impossible de toggle.")
	}
	SetStatus(!status)
	newStatus, err := GetStatusString()
	return newStatus, nil
}

func GetStatusString() (string, error) {
	status, err := GetStatus()
	if err != nil {
		return "", err
	}
	statusString := "fermé"
	if status {
		statusString = "ouvert"
	}
	return statusString, nil
}
