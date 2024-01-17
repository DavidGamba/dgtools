package config

import "fmt"

type Config struct {
	Component map[ID]Component `json:"component"`
	Stack     map[ID]Stack     `json:"stack"`
}

type Component struct {
	ID         ID         `json:"id"`
	Path       string     `json:"path"`
	DependsOn  []string   `json:"depends_on"`
	Variables  []Variable `json:"variables"`
	Workspaces []string   `json:"workspaces"`
}

type ID string

type Variable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Stack struct {
	ID         ID          `json:"id"`
	Components []Component `json:"components"`
}

func (c Component) String() string {
	return fmt.Sprintf("id: %s, path: %s, depends_on: %v, variables: %v, workspaces: %v", c.ID, c.Path, c.DependsOn, c.Variables, c.Workspaces)
}

func (s Stack) String() string {
	componentIDs := make([]ID, len(s.Components))
	for i, c := range s.Components {
		componentIDs[i] = c.ID
	}

	return fmt.Sprintf("id: %s, components: %v", s.ID, componentIDs)
}

func (v Variable) String() string {
	return fmt.Sprintf("%s=%s", v.Name, v.Value)
}
