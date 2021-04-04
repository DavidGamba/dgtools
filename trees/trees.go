package trees

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

var Logger = log.New(ioutil.Discard, "", log.LstdFlags)

// ErrExtraElementsInPath - Indicates when there is a final match and there are remaining path elements.
var ErrExtraElementsInPath = fmt.Errorf("extra elements in path")

// ErrMapKeyNotFound - Key not in config.
var ErrMapKeyNotFound = fmt.Errorf("map key not found")

// ErrNotAnIndex - The given path is not a numerical index and the element is of type slice/array.
var ErrNotAnIndex = fmt.Errorf("not an index")

// ErrInvalidIndex - The given index is invalid.
var ErrInvalidIndex = fmt.Errorf("invalid index")

// NavigateTree allows you to define a path string to traverse a tree composed of maps and arrays.
// To navigate through slices/arrays use a numerical index, for example: [path to array 1]
// When include is true, the returned map will have the key as part of it.
func NavigateTree(include bool, m interface{}, p []string) (interface{}, []string, error) {
	// Logger.Printf("type: %v, path: %v\n", reflect.TypeOf(m), p)
	path := strings.Join(p, "/")
	Logger.Printf("NavigateTree: Self: %v, Input path: '%s'", include, path)
	if len(p) <= 0 {
		return m, p, nil
	}
	switch m.(type) {
	case map[interface{}]interface{}:
		Logger.Printf("NavigateTree: map type")
		t, ok := m.(map[interface{}]interface{})[p[0]]
		if !ok {
			return m, p, fmt.Errorf("%w: %s", ErrMapKeyNotFound, p[0])
		}
		if include && len(p) == 1 {
			Logger.Printf("NavigateTree: self return")
			return map[interface{}]interface{}{p[0]: m.(map[interface{}]interface{})[p[0]]}, p[1:], nil
		}
		return NavigateTree(include, t, p[1:])
	case map[string]interface{}:
		Logger.Printf("NavigateTree: map type")
		t, ok := m.(map[string]interface{})[p[0]]
		if !ok {
			return m, p, fmt.Errorf("%w: %s", ErrMapKeyNotFound, p[0])
		}
		if include && len(p) == 1 {
			Logger.Printf("NavigateTree: self return")
			return map[interface{}]interface{}{p[0]: m.(map[string]interface{})[p[0]]}, p[1:], nil
		}
		return NavigateTree(include, t, p[1:])
	case []interface{}:
		Logger.Printf("NavigateTree: slice/array type")

		index, err := strconv.Atoi(p[0])
		if err != nil {
			return m, p, fmt.Errorf("%w: %s", ErrNotAnIndex, p[0])
		}
		if index < 0 || len(m.([]interface{})) <= index {
			return m, p, fmt.Errorf("%w: %s", ErrInvalidIndex, p[0])
		}
		return NavigateTree(include, m.([]interface{})[index], p[1:])
	default:
		Logger.Printf("NavigateTree: single element type")
		return m, p, fmt.Errorf("%w: %s", ErrExtraElementsInPath, strings.Join(p, "/"))
	}
}

