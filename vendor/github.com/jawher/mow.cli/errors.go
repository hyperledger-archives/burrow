package cli

import (
	"errors"
)

var (
	errHelpRequested    = errors.New("Help requested")
	errVersionRequested = errors.New("Version requested")
)
