package yamlpath

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/caspr-io/yamlpath/parts"
)

// https://pypi.org/project/yamlpath/#supported-yaml-path-segments

const (
	KeyPart           = "^[a-zA-Z0-9_\\.-]+$"
	KeySearchPart     = "^\\[\\.=[a-zA-Z][a-zA-Z0-9_-]*\\]$"
	ExplicitIndexPart = "^\\[[0-9]+\\]$"
	ImplicitIndexPart = "^[0-9]+$"
	SlicePart         = "^\\[[0-9]+:[0-9]+\\]$"
	ValueSearchPart   = "^\\[\\.[\\^\\$\\%].+\\]$"
)

var regexps map[string]*regexp.Regexp = map[string]*regexp.Regexp{ //nolint:gochecknoglobals
	KeyPart:           regexp.MustCompile(KeyPart),
	KeySearchPart:     regexp.MustCompile(KeySearchPart),
	ExplicitIndexPart: regexp.MustCompile(ExplicitIndexPart),
	ImplicitIndexPart: regexp.MustCompile(ImplicitIndexPart),
	SlicePart:         regexp.MustCompile(SlicePart),
	ValueSearchPart:   regexp.MustCompile(ValueSearchPart),
}

// YamlPath traverses the yaml document to and returns the retrieved value
func YamlPath(yaml map[string]interface{}, path string) (interface{}, error) {
	splitPath, err := parts.ParsePath(path)
	if err != nil {
		return nil, PathError(path, err)
	}
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
	case regexps[ExplicitIndexPart].MatchString(part):
		i, err := strconv.Atoi(part[1 : len(part)-1])
		if err != nil {
			return nil, err
		}

		if i < len(l) {
			return l[i], nil
		}

		return nil, fmt.Errorf("out of bounds '%d' for array of length '%d'", i, len(l))
	case regexps[SlicePart].MatchString(part):
		idxs := strings.Split(part[1:len(part)-1], ":")

		start, err := strconv.Atoi(idxs[0])
		if err != nil {
			return nil, fmt.Errorf("part '%s' is not an index into an array. %w", part, err)
		}

		end, err := strconv.Atoi(idxs[1])
		if err != nil {
			return nil, fmt.Errorf("part '%s' is not an index into an array. %w", part, err)
		}

		if start > end {
			return nil, fmt.Errorf("cannot take slice with reversed indexes '%s'", part)
		}

		if start >= len(l) {
			return nil, fmt.Errorf("start slice index out of bounds '%d' for array length '%d'", start, len(l))
		}

		if end >= len(l) {
			return nil, fmt.Errorf("end slice index out of bounds '%d' for array length '%d'", end, len(l))
		}

		slice := []interface{}{}
		for i := start; i < end; i++ {
			slice = append(slice, l[i])
		}

		return slice, nil
	case regexps[ImplicitIndexPart].MatchString(part):
		i, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("part '%s' is not an index into an array. %w", part, err)
		}

		return l[i], nil
	case regexps[KeyPart].MatchString(part):
		result := []interface{}{}

		for _, v := range l {
			r, err := navigateYaml(v, part)
			if err != nil {
				return nil, err
			}

			result = append(result, r)
		}

		return result, nil
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
		if v, ok := m[part]; ok {
			return v, nil
		}

		return nil, fmt.Errorf("could not find key '%s' in yaml", part)
	case regexps[KeySearchPart].MatchString(part):
		key := part[3 : len(part)-1]
		return m[key], nil
	default:
		return nil, fmt.Errorf("no support for part '%s'", part)
	}
}
