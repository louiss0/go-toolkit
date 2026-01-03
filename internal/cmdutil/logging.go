package cmdutil

import (
	"github.com/charmbracelet/log"
	"github.com/louiss0/g-tools/mode"
)

func LogInfoIfProduction(message string, args ...any) {
	mode.NewModeOperator().ExecuteIfModeIsProduction(func() {
		log.Infof(message, args...)
	})
}
