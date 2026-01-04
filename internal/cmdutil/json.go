package cmdutil

import (
	"encoding/json"
	"io"

	"github.com/tidwall/pretty"
)

func WritePrettyJSON(writer io.Writer, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = writer.Write(pretty.Pretty(raw))
	return err
}
