package parts

import "fmt"

type YamlPathPart interface {
	NavigateMap(map[string]interface{}) (interface{}, error)
	NavigateArray([]interface{}) (interface{}, error)
}

func ParsePath(path string) ([]string, error) {
	if path[0] == '/' {
		return parsePath(path[1:], '/')
	}

	return parsePath(path, '.')
}

func parsePath(path string, separator rune) ([]string, error) {
	parts := []string{}
	currentSegment := []rune{}
	i := 0

	for i < len(path) {
		r := rune(path[i])
		switch r {
		case separator:
			if err := addPart(currentSegment, &parts); err != nil {
				return nil, err
			}

			currentSegment = []rune{}
		case '\\':
			i++
			r = rune(path[i])
			currentSegment = append(currentSegment, r)
		case '"', '\'':
			if err := addPart(currentSegment, &parts); err != nil {
				return nil, err
			}

			currentSegment = []rune{}

			p, endIdx, err := parsePathUntil(path, i+1, r, false) //nolint:gomnd
			if err != nil {
				return nil, err
			}

			parts = append(parts, p)
			i = endIdx
		case '[':
			if err := addPart(currentSegment, &parts); err != nil {
				return nil, err
			}

			currentSegment = []rune{}

			p, endIdx, err := parsePathUntil(path, i, ']', true)
			if err != nil {
				return nil, err
			}

			parts = append(parts, p)
			i = endIdx
		default:
			currentSegment = append(currentSegment, r)
		}
		i++
	}

	if len(currentSegment) > 0 {
		if err := addPart(currentSegment, &parts); err != nil {
			return nil, err
		}
	}

	return parts, nil
}

func addPart(part []rune, parts *[]string) error {
	if len(part) == 0 {
		return nil
	}

	p, err := DetectPart(string(part))
	if err != nil {
		return err
	}

	l := append(*parts, p)
	*parts = l

	return nil
}

func DetectPart(s string) (string, error) {
	return s, nil
}

func parsePathUntil(path string, idx int, stopOn rune, inclusive bool) (string, int, error) {
	part := []rune{}
	i := idx

	for i < len(path) {
		r := rune(path[i])
		part = append(part, r)

		if r == stopOn && inclusive {
			return string(part), i + 1, nil
		} else if r == stopOn {
			return string(part[0 : len(part)-1]), i + 1, nil
		}
		i++
	}

	return "", -1, fmt.Errorf("could not find terminating '%c' in path '%s'", stopOn, path[idx:])
}
