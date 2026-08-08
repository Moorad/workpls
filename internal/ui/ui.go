package ui

import (
	"fmt"
	"os"

	"charm.land/huh/v2"
	"golang.org/x/term"
)

type Choice struct {
	Label string
	Value string
}

func RequireTerminal(stdin, stdout *os.File) error {
	if !term.IsTerminal(int(stdin.Fd())) || !term.IsTerminal(int(stdout.Fd())) {
		return fmt.Errorf("Workpls requires an interactive terminal on stdin and stdout")
	}
	return nil
}

func Input(prompt string, validationFunc func(string) error) (string, error) {
	var answer string
	err := huh.NewInput().
		Title(prompt).
		Prompt(">").
		Validate(validationFunc).
		Value(&answer).Run()

	return answer, err
}

func Confirm(prompt string) (bool, error) {
	var confirm bool

	err := huh.NewConfirm().
		Title(prompt).
		Affirmative("Yes").
		Negative("No").
		Value(&confirm).Run()

	if err == huh.ErrUserAborted {
		return false, nil
	}

	return confirm, err
}

func Select(title string, choices []Choice) (value string, selected bool, err error) {
	options := []huh.Option[string]{}

	for _, choice := range choices {
		options = append(options, huh.NewOption(choice.Label, choice.Value))
	}

	var choice string
	err = huh.NewSelect[string]().
		Title(title).
		Options(
			options...,
		).
		Value(&choice).
		Run()

	if err == huh.ErrUserAborted {
		return "", false, nil
	}

	if err != nil {
		return "", false, err
	}

	return choice, true, nil
}
