package httpproxy

import (
	"fmt"
	"strings"
	"testing"
)

func Test_HistoryReader(t *testing.T) {
	r := strings.NewReader("hello")
	hr := HistoryReader{
		reader: r,
	}
	a := make([]byte, 5)
	n, err := hr.Read(a)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(n, a)

	if string(hr.History()) != "hello" {
		t.Fail()
	}
}
