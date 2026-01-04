package prompt

import (
	"errors"

	"github.com/charmbracelet/huh"
	"github.com/louiss0/g-tools/mode"
	"github.com/spf13/cobra"
)

var ErrPromptsDisabled = errors.New("prompts disabled outside production mode")

type Input struct {
	Title       string
	Description string
	Placeholder string
	Validate    func(string) error
}

type Option struct {
	Label string
	Value string
}

type Select struct {
	Title   string
	Options []Option
}

type Runner interface {
	Input(cmd *cobra.Command, input Input) (string, error)
	Select(cmd *cobra.Command, selectInput Select) (string, error)
}

type HuhRunner struct {
	mode mode.ModeOperator
}

func NewRunner(modeOperator mode.ModeOperator) Runner {
	return HuhRunner{mode: modeOperator}
}

func (r HuhRunner) Input(cmd *cobra.Command, input Input) (string, error) {
	if !r.mode.IsProductionMode() {
		return "", ErrPromptsDisabled
	}

	value := ""
	field := huh.NewInput().Title(input.Title).Value(&value)
	if input.Description != "" {
		field.Description(input.Description)
	}
	if input.Placeholder != "" {
		field.Placeholder(input.Placeholder)
	}
	if input.Validate != nil {
		field.Validate(input.Validate)
	}

	if err := runField(cmd, field); err != nil {
		return "", err
	}

	return value, nil
}

func (r HuhRunner) Select(cmd *cobra.Command, selectInput Select) (string, error) {
	if !r.mode.IsProductionMode() {
		return "", ErrPromptsDisabled
	}

	value := ""
	field := huh.NewSelect[string]().Title(selectInput.Title).Value(&value)
	for _, option := range selectInput.Options {
		field.Options(huh.NewOption(option.Label, option.Value))
	}

	if err := runField(cmd, field); err != nil {
		return "", err
	}

	return value, nil
}

func runField(cmd *cobra.Command, field huh.Field) error {
	form := huh.NewForm(huh.NewGroup(field)).
		WithInput(cmd.InOrStdin()).
		WithOutput(cmd.ErrOrStderr())

	return form.Run()
}
