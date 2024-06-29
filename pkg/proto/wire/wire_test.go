package wire_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"hookt.dev/cmd/pkg/proto/wire"
)

func TestParse(t *testing.T) {
	cases := []struct {
		file string
		ok   bool
	}{
		{"../../testdata/ok.yaml", true},
		{"../../testdata/bad/1.yaml", false},
		{"../../testdata/bad/2.yaml", false},
		{"../../testdata/bad/3.yaml", false},
	}

	for _, c := range cases {
		t.Run(filepath.Base(c.file), func(t *testing.T) {
			p := file(t, c.file)

			_, err := wire.XParse(p)
			if c.ok && err != nil {
				t.Fatal(err)
			} else if !c.ok && err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func file(t *testing.T, path string) []byte {
	t.Helper()

	p, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return p
}
