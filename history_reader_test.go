package httpproxy

import (
	"strings"
	"testing"
)

func Test_HistoryReader(t *testing.T) {
	r := strings.NewReader("hello")
	HistoryReader{}
}
