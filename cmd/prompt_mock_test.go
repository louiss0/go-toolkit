package cmd

import (
	"errors"
	"fmt"

	"github.com/louiss0/cobra-cli-template/internal/prompt"
	"github.com/spf13/cobra"
)

type promptStepKind string

const (
	promptStepInput  promptStepKind = "input"
	promptStepSelect promptStepKind = "select"
)

type promptStep struct {
	kind  promptStepKind
	value string
	err   error
}

type promptMock struct {
	steps []promptStep
	index int
}

func newPromptMock(steps ...promptStep) *promptMock {
	return &promptMock{steps: steps}
}

func (m *promptMock) Input(_ *cobra.Command, input prompt.Input) (string, error) {
	step, err := m.next(promptStepInput)
	if err != nil {
		return "", err
	}
	if step.err != nil {
		return "", step.err
	}
	if input.Validate != nil {
		if err := input.Validate(step.value); err != nil {
			return "", err
		}
	}
	return step.value, nil
}

func (m *promptMock) Select(_ *cobra.Command, selectInput prompt.Select) (string, error) {
	step, err := m.next(promptStepSelect)
	if err != nil {
		return "", err
	}
	if step.err != nil {
		return "", step.err
	}
	if !containsOption(selectInput.Options, step.value) {
		return "", fmt.Errorf("unexpected selection: %s", step.value)
	}
	return step.value, nil
}

func (m *promptMock) next(expected promptStepKind) (promptStep, error) {
	if m.index >= len(m.steps) {
		return promptStep{}, errors.New("prompt mock: no steps remaining")
	}
	step := m.steps[m.index]
	m.index++
	if step.kind != expected {
		return promptStep{}, fmt.Errorf("prompt mock: expected %s, got %s", expected, step.kind)
	}
	return step, nil
}

func containsOption(options []prompt.Option, value string) bool {
	for _, option := range options {
		if option.Value == value {
			return true
		}
	}
	return false
}
