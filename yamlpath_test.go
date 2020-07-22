package yamlpath

import (
	"testing"

	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
)

//nolint:gochecknoglobals
var yamlDoc = `---
hash:
  child_attr:
    key: 5280
  dotted.child:
    key: 42
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
:f:oo::
  bar: baz
`

var tests = []struct { //nolint:gochecknoglobals
	testName string
	path     string
	output   interface{}
}{
	{"DotNotation", "hash.child_attr.key", 5280},
	{"DotNotation", ":f:oo:.bar", "baz"},
	{"SlashNotation", "/hash/child_attr/key", 5280},
	{"EscapedDotNotation", "hash.dotted\\.child.key", 42},
	{"QuotedDotNotation", "hash.\"dotted.child\".key", 42},
	{"SingleQuotedDotNotation", "hash.'dotted.child'.key", 42},
	{"SlashDotted", "/hash/dotted.child/key", 42},
	{"SearchChildKeyDot", "hash.child_attr[.=key]", 5280},
	{"SearchChildKeySlash", "/hash/child_attr[.=key]", 5280},
	{"ExplicitIndex", "aliases[0]", "Simple string value"},
	{"ImplicitIndex", "aliases.0", "Simple string value"},
	{"ArraySlice", "aliases[0:2]", []interface{}{"Simple string value", "Complex ending"}},
	{"ValuePrefix", "aliases[.^Simple]", "Simple string value"},
	{"ValueContains", "aliases[.%string]", "Simple string value"},
	{"ValueSuffix", "aliases[.$value]", "Simple string value"},
	// TODO implement regex support
	// {"ValueRegex", "aliases[.=~/^(\\b[Ss][a-z]+\\s){2}[a-z]+$/]", "Simple string value"},
	{"SlashExplicitIndex", "/aliases[0]", "Simple string value"},
	{"SlashImplicitIndex", "/aliases/0", "Simple string value"},
	{"GetArrayOfHashes", "/users/name", []interface{}{"User One", "User Two"}},
	{"IndexIntoArrayOfHashes", "/users[1]/name", "User Two"},
	// Unsupported as anchors are erased in parsed yaml
	// {"AnchoredIndex", "aliases[&first_anchor]", "Simple string value"},
	// {"SlashAnchoredIndex", "/aliases[&first_anchor]", "Simple string value"},
	{"ErrorOnNonExistingKey", "/broken", nil},
	{"ErrorOnNonExistingArrayIndex", "aliases[4]", nil},
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
			if expected != nil {
				assert.NilError(t, err)
				assert.DeepEqual(t, out, expected)
			} else {
				assert.Assert(t, err != nil)
			}
		})
	}
}
