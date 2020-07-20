// Package cfn provides the Template type that models a CloudFormation template.
//
// The sub-packages of cfn contain various tools for working with templates
package cfn

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/aws-cloudformation/rain/cfn/diff"
	"github.com/aws-cloudformation/rain/cfn/graph"
)

const pseudoParameterType = "Parameter"

// Element represents a top-level entry in a CloudFormation template
// for example a resource, parameter, or output
type Element struct {
	// Name is the name of the element
	Name string

	// Type is the name of the top-level part of a CloudFormation
	// that contains this Element (e.g. Resources, Parameters)
	Type string
}

// Template represents a CloudFormation template. The Template type
// is minimal for now but will likely grow new features as needed by rain.
type Template struct {
	yaml.Node
}

// NewTemplate returns a Template constructed from the provided map
func NewTemplate(in map[string]interface{}) Template {
	t := Template{}

	err := t.Encode(in)
	if err != nil {
		panic(fmt.Errorf("Error converting map to template: %s", err))
	}

	return t
}

func getNode(node *yaml.Node, path []interface{}) (*yaml.Node, error) {
	if len(path) == 0 {
		return node, nil
	}

	index, path := path[0], path[1:]

	switch node.Kind {
	case yaml.DocumentNode, yaml.MappingNode:
		s, ok := index.(string)
		if !ok {
			return nil, fmt.Errorf("Attempted to index a map with a %T", index)
		}

		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Value == s {
				return getNode(node.Content[i+1], path)
			}
		}

		return nil, fmt.Errorf("Unable to find map key '%s'", s)
	case yaml.SequenceNode:
		i, ok := index.(int)
		if !ok {
			return nil, fmt.Errorf("Attmpted to index a sequence with a %T", index)
		}

		if i < 0 || i >= len(node.Content) {
			return nil, fmt.Errorf("Sequence index out of range: %d", i)
		}

		return getNode(node.Content[i], path)
	case yaml.ScalarNode:
		return nil, errors.New("Attempted to index a scalar")
	case yaml.AliasNode:
		return nil, errors.New("Alias nodes are not supported")
	default:
		return nil, fmt.Errorf("Unexpected node Kind: %d", node.Kind)
	}
}

func setNode(node *yaml.Node, path []interface{}, value interface{}) error {
	if len(path) == 0 {
		return node.Encode(value)
	}

	index, nextPath := path[0], path[1:]

	switch node.Kind {
	case yaml.DocumentNode, yaml.MappingNode:
		s, ok := index.(string)
		if !ok {
			return fmt.Errorf("Attempted to index a map with a %T", index)
		}

		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Value == s {
				return setNode(node.Content[i+1], nextPath, value)
			}
		}

		// Create a new entry in the map
		keyNode := yaml.Node{}
		keyNode.SetString(s)

		valueNode := yaml.Node{}
		err := setNode(&valueNode, nextPath, value)
		if err != nil {
			return err
		}

		node.Content = append(node.Content, &keyNode, &valueNode)

		return nil
	case yaml.SequenceNode:
		i, ok := index.(int)
		if !ok {
			return fmt.Errorf("Attmpted to index a sequence with a %T", index)
		}

		if i < 0 || i > len(node.Content) {
			return fmt.Errorf("Sequence index out of range: %d", i)
		}

		if i < len(node.Content) {
			return setNode(node.Content[i], nextPath, value)
		}

		// Create a new entry in the sequence
		valueNode := yaml.Node{}
		err := setNode(&valueNode, nextPath, value)
		if err != nil {
			return err
		}

		node.Content = append(node.Content, &valueNode)

		return nil
	case yaml.ScalarNode:
		return errors.New("Attempted to index a scalar")
	case yaml.AliasNode:
		return errors.New("Alias nodes are not supported")
	default:
		// We're creating a new thing
		switch index.(type) {
		case string:
			err := node.Encode(map[string]interface{}{})
			if err != nil {
				return err
			}

			return setNode(node, path, value)
		case int:
			err := node.Encode([]interface{}{})
			if err != nil {
				return err
			}

			return setNode(node, path, value)
		default:
			return fmt.Errorf("Unexpected index element: %#v", index)
		}
	}
}

// Get returns the value at `path`.
// An error is returned if there is no value at the given path
// or the path is inaccessible.
func (t Template) Get(path []interface{}) (interface{}, error) {
	return getNode(&t.Node, path)
}

// Set set the value at `path` to `value`.
// An error is returned if the path is inaccessible
func (t *Template) Set(path []interface{}, value interface{}) error {
	return setNode(&t.Node, path, value)
}

// Map returns the template as a map[string]interface{}
func (t Template) Map() map[string]interface{} {
	var out map[string]interface{}

	err := t.Decode(&out)
	if err != nil {
		panic(fmt.Errorf("Error converting template to map: %s", err))
	}

	return out
}

// Diff returns a Diff object representing the difference
// between this template and the template passed to Diff
func (t Template) Diff(other Template) diff.Diff {
	return diff.New(t.Map(), other.Map())
}

// Graph returns a Graph representing the connections
// between elements in the template.
// The type of each item in the graph should be Element
func (t Template) Graph() graph.Graph {
	// Map out parameter and resource names so we know which is which
	entities := make(map[string]string)
	for typeName, entity := range t.Map() {
		if typeName != "Parameters" && typeName != "Resources" {
			continue
		}

		if entityTree, ok := entity.(map[string]interface{}); ok {
			for entityName := range entityTree {
				entities[entityName] = typeName
			}
		}
	}

	// Now find the deps
	graph := graph.New()
	for typeName, entity := range t.Map() {
		if typeName != "Resources" && typeName != "Outputs" {
			continue
		}

		if entityTree, ok := entity.(map[string]interface{}); ok {
			for fromName, res := range entityTree {
				from := Element{fromName, typeName}
				graph.Add(from)

				resource := res.(map[string]interface{})
				for _, toName := range getRefs(resource) {
					toName = strings.Split(toName, ".")[0]

					toType, ok := entities[toName]

					if !ok {
						if strings.HasPrefix(toName, "AWS::") {
							toType = "Parameters"
						} else {
							panic(fmt.Sprintf("Template has unresolved dependency '%s' at %s: %s", toName, typeName, fromName))
						}
					}

					graph.Add(from, Element{toName, toType})
				}
			}
		}
	}

	return graph
}
