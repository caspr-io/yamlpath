package yamlpath

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	KeyPart         = "^[a-zA-Z0-9_-]+$"
	KeySearchPart   = "^\\[\\.=[a-zA-Z][a-zA-Z0-9_-]*\\]$"
	IndexPart       = "^\\[[0-9]+\\]$"
	ValueSearchPart = "^\\[\\.[\\^\\$\\%].+\\]$"
)

var regexps map[string]*regexp.Regexp = map[string]*regexp.Regexp{ //nolint:gochecknoglobals
	KeyPart:         regexp.MustCompile(KeyPart),
	KeySearchPart:   regexp.MustCompile(KeySearchPart),
	IndexPart:       regexp.MustCompile(IndexPart),
	ValueSearchPart: regexp.MustCompile(ValueSearchPart),
}

// YamlPath traverses the yaml document to and returns the retrieved value
func YamlPath(yaml map[string]interface{}, path string) (interface{}, error) {
	splitPath := parsePath(path)
	// fmt.Printf("%v, %d", splitPath, len(splitPath))

	var value interface{} = yaml

	for _, pathPart := range splitPath {
		returned, err := navigateYaml(value, pathPart)
		if err != nil {
			return nil, PathError(path, err)
		}

		value = returned
	}

	return value, nil
}

func navigateYaml(yaml interface{}, part string) (interface{}, error) {
	switch y := yaml.(type) {
	case map[string]interface{}:
		return navigateMap(y, part)
	case []interface{}:
		return navigateArray(y, part)
	default:
		return nil, fmt.Errorf("no support yet for %v", yaml)
	}
}

func navigateArray(l []interface{}, part string) (interface{}, error) {
	switch {
	case regexps[IndexPart].MatchString(part):
		i, err := strconv.Atoi(part[1 : len(part)-1])
		if err != nil {
			return nil, err
		}

		return l[i], nil
	case regexps[KeyPart].MatchString(part):
		i, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("part '%s' is not an index into an array. %w", part, err)
		}

		return l[i], nil
	case regexps[ValueSearchPart].MatchString(part):
		toFind := part[3 : len(part)-1]
		operator := part[2]
		for _, i := range l {
			switch s := i.(type) {
			case string:
				if valueMatches(s, toFind, operator) {
					return s, nil
				}
				continue
			default:
				return nil, fmt.Errorf("could not search for value '%s' as list does not contain strings", part)
			}
		}
		return nil, fmt.Errorf("could not find match for search part '%s'", part)
	default:
		return nil, fmt.Errorf("part '%s' not supported for array", part)
	}
}

func valueMatches(s string, find string, operator byte) bool {
	switch operator {
	case '^':
		return strings.HasPrefix(s, find)
	case '$':
		return strings.HasSuffix(s, find)
	case '%':
		return strings.Contains(s, find)
	default:
		return false
	}
}

func navigateMap(m map[string]interface{}, part string) (interface{}, error) {
	switch {
	case regexps[KeyPart].MatchString(part):
		return m[part], nil
	case regexps[KeySearchPart].MatchString(part):
		key := part[3 : len(part)-1]
		return m[key], nil
	default:
		return nil, fmt.Errorf("no support for part '%s'", part)
	}
}

func parsePath(path string) []string {
	parts := []string{}
	current := []rune{}
	i := 0
	var r rune
	for i < len(path) {
		r = rune(path[i])
		switch r {
		case '.':
			parts = append(parts, string(current))
			current = []rune{}
		case '/':
			if len(current) > 0 {
				parts = append(parts, string(current))
			}

			current = []rune{}
		case '[':
			if len(current) > 0 {
				parts = append(parts, string(current))
				current = []rune{}
			}
			for i < len(path) {
				current = append(current, r)
				if path[i] == ']' {
					break
				}
				i++
				r = rune(path[i])
			}
		default:
			current = append(current, r)
		}
		i++
	}

	if len(current) > 0 {
		parts = append(parts, string(current))
	}

	return parts
}
