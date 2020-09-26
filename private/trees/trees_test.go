package trees

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestNavigateTree(t *testing.T) {
	tests := []struct {
		name         string
		include      bool
		path         []string
		input        interface{}
		expected     interface{}
		expectedPath []string
		err          error
	}{
		{"string config, no path", false,
			[]string{}, "hola", "hola", []string{}, nil},
		{"string config, bad path", false,
			[]string{"extra", "elements"}, "hola", "hola", []string{"extra", "elements"}, ErrExtraElementsInPath},
		{"array config, no path", false,
			[]string{}, []interface{}{"one", "two", "three"}, []interface{}{"one", "two", "three"}, []string{}, nil},
		{"array config, path", false,
			[]string{"1"}, []interface{}{"one", "two", "three"}, "two", []string{}, nil},
		{"array config, bad path", false,
			[]string{"[1]"}, []interface{}{"one", "two", "three"}, []interface{}{"one", "two", "three"}, []string{"[1]"}, ErrNotAnIndex},
		{"map config, no path", false,
			[]string{},
			map[interface{}]interface{}{"map": []string{"one", "two", "three"}},
			map[interface{}]interface{}{"map": []string{"one", "two", "three"}},
			[]string{}, nil},
		{"map config, path", false,
			[]string{"map"},
			map[interface{}]interface{}{"map": []string{"one", "two", "three"}, "another": []string{"four", "five", "six"}},
			[]string{"one", "two", "three"},
			[]string{}, nil},
		{"map config, path", true,
			[]string{"map"},
			map[interface{}]interface{}{"map": []string{"one", "two", "three"}, "another": []string{"four", "five", "six"}},
			map[interface{}]interface{}{"map": []string{"one", "two", "three"}},
			[]string{}, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := ""
			buf := bytes.NewBufferString(s)
			Logger.SetOutput(buf)
			output, path, err := NavigateTree(test.include, test.input, test.path)
			if !errors.Is(err, test.err) {
				t.Errorf("Unexpected error: %s\n", err)
			}
			if !reflect.DeepEqual(path, test.expectedPath) {
				t.Errorf("Expected:\n%#v\nGot:\n%#v\n", test.expectedPath, path)
			}
			if !reflect.DeepEqual(output, test.expected) {
				t.Errorf("Expected:\n%#v\nGot:\n%#v\n", test.expected, output)
			}
			t.Log(buf.String())
		})
	}
}
