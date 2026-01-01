// Package custom_errors provides error handling functionality for flag-related and argument-related operations.
package custom_errors

import (
	"errors"
	"fmt"
	"regexp"
)

// InvalidFlag represents an error indicating an invalid flag.
var InvalidFlag = errors.New("Invalid Flag:")

// InvalidArgument represents an error indicating an invalid argument.
var InvalidArgument = errors.New("Invalid Argument:")

// FlagName is a string type representing the name of a flag.
type FlagName string

// Error validates the FlagName and returns an error if it's invalid.
// A valid flag name must contain only alphanumeric characters.
func (self FlagName) Error() error {

	regex := regexp.MustCompile(`^[a-z0-9]+$`)

	if !regex.MatchString(string(self)) {
		return fmt.Errorf("%w %s a flag name must be alphanumeric from start to end %s", InvalidFlag, self, string(self))
	}

	return nil
}

// CreateInvalidFlagErrorWithMessage creates an error with a custom message for an invalid flag.
// It first validates the flag name and returns the validation error if present.
var CreateInvalidFlagErrorWithMessage = func(flagName FlagName, message string) error {

	if err := flagName.Error(); err != nil {
		return err
	}

	return fmt.Errorf("%w %s %s", InvalidFlag, flagName, message)

}

// CreateInvalidArgumentErrorWithMessage creates an error with a custom message for an invalid argument.
var CreateInvalidArgumentErrorWithMessage = func(message string) error {

	return fmt.Errorf("%w %s", InvalidFlag, message)

}
