// Package custom_flags provides custom flag types for command-line argument parsing.
// It implements various flag types that can be used with the cobra CLI framework.
package custom_flags

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/louiss0/cobra-cli-template/custom_errors"
	"github.com/samber/lo"
)

// emptyStringFlag represents a flag that cannot be empty or contain only whitespace
type emptyStringFlag struct {
	value    string
	flagName string
}

// NewEmptyStringFlag creates a new emptyStringFlag with the given flag name
func NewEmptyStringFlag(flagName string) emptyStringFlag {
	return emptyStringFlag{
		flagName: flagName,
	}
}

// String returns the flag's value as a string
func (t emptyStringFlag) String() string {
	return t.value
}

// Set validates and sets the flag's value, checking for empty/whitespace
func (t *emptyStringFlag) Set(value string) error {

	match, error := regexp.MatchString(`^\s+$`, value)

	if error != nil {
		return error
	}

	if match {
		return fmt.Errorf(
			"The %s is empty",
			t.flagName,
		)
	}

	t.value = value
	return nil
}

// Type returns the flag type as a string
func (t emptyStringFlag) Type() string {
	return "string"
}

// boolFlag represents a flag that must be either "true" or "false"
type boolFlag struct {
	value    string
	flagName string
}

// NewBoolFlag creates a new boolFlag with the given flag name
func NewBoolFlag(flagName string) boolFlag {
	return boolFlag{
		flagName: flagName,
	}
}

// String returns the flag's value as a string
func (c boolFlag) String() string {
	return c.value
}

// Set validates and sets the flag's value, ensuring it's a valid boolean
func (c *boolFlag) Set(value string) error {

	match, error := regexp.MatchString(`^\S+$`, value)

	if error != nil {
		return error
	}

	if match && !lo.Contains([]string{"true", "false"}, value) {
		return fmt.Errorf(
			"%sflag must be one of: %v",
			custom_errors.FlagName(c.flagName),
			[]string{"true", "false"},
		)
	}
	c.value = value
	return nil
}

// Type returns the flag type as a string
func (c boolFlag) Type() string {
	return "bool"
}

// Value returns the flag's value as a bool
func (c boolFlag) Value() bool {
	value, _ := strconv.ParseBool(c.value)
	return value
}

// unionFlag represents a flag that must be one of a predefined set of values
type unionFlag struct {
	value         string
	allowedValues []string
	flagName      string
}

// NewUnionFlag creates a new unionFlag with the given allowed values and flag name
func NewUnionFlag(allowedValues []string, flagName string) unionFlag {
	return unionFlag{
		allowedValues: allowedValues,
		flagName:      flagName,
	}
}

// String returns the flag's value as a string
func (self unionFlag) String() string {
	return self.value
}

// Set validates and sets the flag's value, ensuring it's one of the allowed values
func (self *unionFlag) Set(value string) error {

	match, error := regexp.MatchString(`^\S+$`, value)

	if error != nil {
		return error
	}

	if match && !lo.Contains(self.allowedValues, value) {
		return fmt.Errorf(
			"%sflag must be one of: %v",
			custom_errors.FlagName(self.flagName),
			self.allowedValues,
		)

	}
	self.value = value
	return nil
}

// Type returns the flag type as a string
func (self unionFlag) Type() string {
	return "string"
}

// RangeFlag represents a flag that must be an integer within a specified range
type RangeFlag struct {
	value, min, max int
	flagName        string
}

// NewRangeFlag creates a new RangeFlag with the given flag name and range bounds
func NewRangeFlag(flagName string, min, max int) RangeFlag {

	if min > max {
		panic("min must be less than max")
	}

	if min < 0 || max < 0 {
		panic("min and max must be non-negative")
	}

	if min > max {
		panic("min must be less than max")
	}

	if min < 0 || max < 0 {
		panic("min and max must be non-negative")
	}

	return RangeFlag{
		min:      min,
		max:      max,
		flagName: flagName,
	}
}

// String returns the flag's value as a string
func (self RangeFlag) String() string {
	return fmt.Sprintf("%d", self.value)
}

// Value returns the flag's value as an int
func (self RangeFlag) Value() int {
	return self.value
}

// Set validates and sets the flag's value, ensuring it's within the allowed range
func (self *RangeFlag) Set(value string) error {

	match, error := regexp.MatchString(`^\d+$`, value)

	if error != nil {
		return error
	}

	if match {
		num, _ := strconv.Atoi(value)
		if num < self.min || num > self.max {
			return fmt.Errorf(
				"%sflag must be between %d and %d",
				custom_errors.FlagName(self.flagName),
				self.min,
				self.max,
			)
		}
		self.value = num
		return nil
	}

	return fmt.Errorf(
		"%sflag must be an integer between %d and %d",
		custom_errors.FlagName(self.flagName),
		self.min,
		self.max,
	)
}

// Type returns the flag type as a string
func (self RangeFlag) Type() string {
	return "string"
}
