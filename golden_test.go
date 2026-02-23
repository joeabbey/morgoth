package morgoth_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joeabbey/morgoth/internal/eval"
	"github.com/joeabbey/morgoth/internal/lexer"
	"github.com/joeabbey/morgoth/internal/parser"
)

func TestGoldenExamples(t *testing.T) {
	examples, err := filepath.Glob("examples/*.mor")
	if err != nil {
		t.Fatal(err)
	}
	if len(examples) == 0 {
		t.Fatal("no example files found")
	}

	for _, exFile := range examples {
		name := strings.TrimSuffix(filepath.Base(exFile), ".mor")
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile(exFile)
			if err != nil {
				t.Fatalf("failed to read %s: %v", exFile, err)
			}

			goldenFile := filepath.Join("testdata", name+".golden")
			expected, err := os.ReadFile(goldenFile)
			if err != nil {
				t.Fatalf("failed to read golden file %s: %v", goldenFile, err)
			}

			l := lexer.New(string(source))
			p := parser.New(l)
			program := p.Parse()
			if errs := p.Errors(); len(errs) > 0 {
				t.Fatalf("parse errors: %v", errs)
			}

			var buf bytes.Buffer
			e := eval.New()
			e.SetOutput(&buf)
			_, evalErr := e.Eval(program)
			if evalErr != nil {
				t.Fatalf("eval error: %v", evalErr)
			}

			got := buf.String()
			want := string(expected)
			if got != want {
				t.Errorf("output mismatch for %s:\ngot:  %q\nwant: %q", exFile, got, want)
			}
		})
	}
}
