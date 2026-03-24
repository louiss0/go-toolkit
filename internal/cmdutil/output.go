package cmdutil

import (
	"fmt"
	"io"
)

func WriteLine(writer io.Writer, value string) error {
	_, err := fmt.Fprintln(writer, value)
	return err
}
