package httpproxy

import (
	"io"
	"regexp"
)

var hasPort = regexp.MustCompile(`:\d+$`)
