package cueutils_test

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/DavidGamba/dgtools/cueutils"
)

func setupLogging() *bytes.Buffer {
	s := ""
	buf := bytes.NewBufferString(s)
	cueutils.Logger.SetOutput(buf)
	return buf
}

func TestUnmarshal(t *testing.T) {
	// Given
	tests := []struct {
		name       string
		p          string
		files      []string
		schema     string
		data       any
		moduleName string
		expected   any
	}{
		{
			name:     "package file1 from dir",
			p:        "file1",
			schema:   "testschemas/file1-schema.cue",
			data:     struct{}{},
			expected: map[string]any{"en": "hello", "es": "hola"},
		},
		{
			name:     "package file1 from overlay",
			p:        "file1",
			files:    []string{"testdata/file1.cue"},
			schema:   "testschemas/file1-schema.cue",
			data:     struct{}{},
			expected: map[string]any{"en": "hello", "es": "hola"},
		},
		{
			name:     "package file2 from dir",
			p:        "file2",
			schema:   "testschemas/file2-schema.cue",
			data:     struct{}{},
			expected: map[string]any{"a": "hello", "b": "hola", "en": "hello", "es": "hola"},
		},
		{
			name:     "package file2 from overlay",
			p:        "file2",
			files:    []string{"testdata/file2.cue"},
			schema:   "testschemas/file2-schema.cue",
			data:     struct{}{},
			expected: map[string]any{"a": "hello", "b": "hola"},
		},
		{
			name:     "package file2 from hidden overlay",
			p:        "file2",
			files:    []string{"testdata/.file2.cue"},
			schema:   "testschemas/file2-schema.cue",
			data:     struct{}{},
			expected: map[string]any{"a": "hello", "b": "hola", "c": "hello", "d": "hola"},
		},
		{
			name:     "no package from dir",
			p:        "_",
			schema:   "testschemas/file3-schema.cue",
			data:     struct{}{},
			expected: map[string]any{"a": "hello", "b": "hola", "en": "hello", "es": "hola"},
		},
		{
			name:     "no package from overlay",
			p:        "_",
			files:    []string{"testdata/file3-b.cue"},
			schema:   "testschemas/file3-schema.cue",
			data:     struct{}{},
			expected: map[string]any{"en": "hello", "es": "hola"},
		},
		{
			name:     "no package from overlay",
			p:        "_",
			files:    []string{"testdata/file3.cue", "testdata/file3-b.cue"},
			schema:   "testschemas/file3-schema.cue",
			data:     struct{}{},
			expected: map[string]any{"a": "hello", "b": "hola", "en": "hello", "es": "hola"},
		},
		{
			name:     "package file4 from dir",
			p:        "file4",
			schema:   "testschemas/file4-schema.cue",
			data:     struct{}{},
			expected: map[string]any{"en": "hello", "es": "hola", "a": "hello", "b": "hola"},
		},
		{
			name:     "package file4 from overlay",
			p:        "file4",
			files:    []string{"testdata/file4.cue"},
			schema:   "testschemas/file4-schema.cue",
			data:     struct{}{},
			expected: map[string]any{"en": "hello", "es": "hola"},
		},
		{
			name:       "package file5 from dir using embed",
			p:          "file5",
			schema:     "testschemas/file5-schema.cue",
			data:       struct{}{},
			moduleName: "file5.cue", // required for embed
			expected:   map[string]any{"en": "hello", "es": "hola"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := setupLogging()

			configs := []cueutils.CueConfigFile{}
			schemaFH, err := os.Open(tt.schema)
			if err != nil {
				t.Log(buf.String())
				t.Fatalf("failed to open '%s': %v", tt.schema, err)
			}
			defer schemaFH.Close()
			configs = append(configs, cueutils.CueConfigFile{Data: schemaFH, Name: tt.schema})
			for _, file := range tt.files {
				configFH, err := os.Open(file)
				if err != nil {
					t.Log(buf.String())
					t.Fatalf("failed to open '%s': %v", file, err)
				}
				defer configFH.Close()
				configs = append(configs, cueutils.CueConfigFile{Data: configFH, Name: file})
			}

			dir := "testdata/"
			if len(tt.files) > 0 {
				dir = ""
			}

			value := cueutils.NewValue()
			err = cueutils.Unmarshal(configs, dir, tt.p, tt.moduleName, value, &tt.data)
			t.Logf("value:\n%#v\n", value)
			if err != nil {
				t.Log(buf.String())
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(tt.data, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, tt.data)
			}

			t.Log(buf.String())
		})
	}
}
