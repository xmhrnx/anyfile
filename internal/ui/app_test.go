package ui

import (
	"reflect"
	"testing"
)

func TestFileArgs(t *testing.T) {
	for _, c := range []struct {
		args, want []string
	}{
		{[]string{"a.zip", "--flag", "b.pdf"}, []string{"a.zip", "b.pdf"}},
		{[]string{"-Embedding"}, nil},
		{[]string{"/Embedding"}, nil},
		{[]string{"/embedding", "a.zip"}, []string{"a.zip"}},
		{[]string{"--giode-action=inspect", "a.zip", "b.pdf"}, nil},
	} {
		if got := fileArgs(c.args); !reflect.DeepEqual(got, c.want) {
			t.Errorf("fileArgs(%q) = %q, want %q", c.args, got, c.want)
		}
	}
}
