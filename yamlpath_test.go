package yamlpath

import (
	"testing"

	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
)

var yamlDoc = `---
hash:
  child_attr:
    key: 5280
aliases:
  - &first_anchor Simple string value
  - Complex ending
users:
  - name: User One
    password: foobar
    roles:
      - Writers
  - name: User Two
    password: barfoo
    roles:
      - Power Users
      - Editors
`

var tests = []struct { //nolint:gochecknoglobals
	testName string
	path     string
	output   interface{}
}{
	{"DotNotation", "hash.child_attr.key", 5280},
	{"SlashNotation", "/hash/child_attr/key", 5280},
	{"SearchChildKeyDot", "hash.child_attr[.=key]", 5280},
	{"SearchChildKeySlash", "/hash/child_attr[.=key]", 5280},
	{"ExplicitIndex", "aliases[0]", "Simple string value"},
	{"ImplicitIndex", "aliases.0", "Simple string value"},
	{"ValuePrefix", "aliases[.^Simple]", "Simple string value"},
	{"ValueContains", "aliases[.%string]", "Simple string value"},
	{"ValueSuffix", "aliases[.$value]", "Simple string value"},
	{"ValueRegex", "aliases[.=~/^(\\b[Ss][a-z]+\\s){2}[a-z]+$/]", "Simple string value"},
	{"SlashExplicitIndex", "/aliases[0]", "Simple string value"},
	{"SlashImplicitIndex", "/aliases/0", "Simple string value"},
	// Unsupported as anchors are erased in parsed yaml
	// {"AnchoredIndex", "aliases[&first_anchor]", "Simple string value"},
	// {"SlashAnchoredIndex", "/aliases[&first_anchor]", "Simple string value"},
}

func TestYamlPath(t *testing.T) {
	for _, tst := range tests {
		path := tst.path
		expected := tst.output
		t.Run(tst.testName, func(t *testing.T) {
			yamlBytes := []byte(yamlDoc)
			yamlMap := map[string]interface{}{}
			assert.NilError(t, yaml.Unmarshal(yamlBytes, &yamlMap))
			out, err := YamlPath(yamlMap, path)
			assert.NilError(t, err)
			assert.Equal(t, out, expected)
		})
	}
}
