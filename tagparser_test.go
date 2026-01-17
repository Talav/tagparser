package tagparser

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type M = map[string]string

var errSimulated = errors.New("simulated error")

func TestParse_EmptyString(t *testing.T) {
	// Test that Parse("") returns empty Options map, not map["":""]
	tag, err := Parse("")
	require.NoError(t, err)
	assert.Equal(t, "", tag.Name)
	assert.Empty(t, tag.Options, "empty string should result in empty Options map")
}

func TestParse_OptionsMode(t *testing.T) {
	// Test Parse (options mode) - all items are options
	tag, err := Parse(`foo,bar=baz`)
	require.NoError(t, err)
	assert.Equal(t, "", tag.Name)
	assert.Equal(t, M{"foo": "", "bar": "baz"}, tag.Options)

	// First item with equals is also an option
	tag2, err := Parse(`foo=bar,baz`)
	require.NoError(t, err)
	assert.Equal(t, "", tag2.Name)
	assert.Equal(t, M{"foo": "bar", "baz": ""}, tag2.Options)

	// All items treated as options
	tag3, err := Parse(`required,email,min=5`)
	require.NoError(t, err)
	assert.Equal(t, "", tag3.Name)
	assert.Equal(t, M{"required": "", "email": "", "min": "5"}, tag3.Options)
}

func TestParseFunc_OptionsMode(t *testing.T) {
	// Test ParseFunc (options mode) - all items are options, no empty keys
	opts := make(M)
	err := ParseFunc(`foo,bar=baz`, func(key, value string) error {
		opts[key] = value

		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, M{"foo": "", "bar": "baz"}, opts)
}

func TestParseWithName(t *testing.T) {
	tests := []struct {
		testName string
		tag      string
		name     string
		opts     map[string]string
		error    string
	}{
		{`empty`, ``, "", nil, ``},

		{`simple 1`, `alfa`, `alfa`, nil, ``},
		{`simple 2`, `alfa,bravo`, `alfa`, M{"bravo": ""}, ``},

		{`quoted key 1`, `'alfa,bravo'`, `alfa,bravo`, nil, ``},
		{`quoted key 2`, `'alfa=bravo'`, `alfa=bravo`, nil, ``},
		{`quoted key 3`, `'alfa\=bravo'`, `alfa=bravo`, nil, ``},
		{`quoted key 4`, "'alfa=bravo'", `alfa=bravo`, nil, ``},

		{`escaped key 1`, `\ =alfa`, "", M{" ": "alfa"}, ""},
		{`escaped key 2`, `' '=alfa`, "", M{" ": "alfa"}, ""},

		{`no name 1`, `,alfa`, "", M{"alfa": ""}, ``},
		{`no name 2`, `,alfa,bravo`, "", M{"alfa": "", "bravo": ""}, ``},
		{`key with empty value`, `alfa=`, "", M{"alfa": ""}, ``},
		{`key with empty quoted value`, `alfa=''`, "", M{"alfa": ""}, ``},
		{`key-value 1`, `alfa=bravo`, "", M{"alfa": "bravo"}, ``},
		{`key-value 2`, `alfa=bravo,charlie`, "", M{"alfa": "bravo", "charlie": ""}, ``},
		{`key-value 3`, `alfa=bravo,charlie=delta`, "", M{"alfa": "bravo", "charlie": "delta"}, ``},

		{`whitespace 1`, `  alfa  `, "alfa", nil, ``},
		{`whitespace 2`, ` alfa ,  bravo  `, "alfa", M{"bravo": ""}, ``},
		{`whitespace 3`, ` alfa, charlie= delta `, "alfa", M{"charlie": "delta"}, ``},

		{`skipped key`, `alfa,,charlie`, "alfa", M{"charlie": ""}, ``},

		{`quoted value 1`, `alfa='bravo,charlie'`, "", M{"alfa": "bravo,charlie"}, ``},
		{`quoted value 2`, `alfa='bravo,charlie',delta`, "", M{"alfa": "bravo,charlie", "delta": ""}, ``},
		{`quoted value 3`, `alfa='bravo=charlie',delta`, "", M{"alfa": "bravo=charlie", "delta": ""}, ``},
		{`quoted value 4`, `alfa='d\'Elta', bravo=charlie`, "", M{"alfa": "d'Elta", "bravo": "charlie"}, ``},

		{`disallowed quote in the middle 1`, `alfa=bravo', charlie 'delta`, "", nil, `quotes must enclose the entire value (at 11)`},
		{`disallowed quote in the middle 2`, `alfa='bravo 'charlie' delta'`, "", nil, `quotes must enclose the entire value (at 13)`},
		{`disallowed quote in the middle of name`, `bravo' charlie'`, "", nil, `quotes must enclose the entire value (at 6)`},
		{`disallowed quote in the middle of name with options`, `alfa,bravo' charlie'`, "alfa", nil, `quotes must enclose the entire value (at 11)`},
		{`disallowed quote in the middle of key`, `bravo' charlie'= delta`, "", nil, `quotes must enclose the entire value (at 6)`},

		{`disallowed vmihailenco-style parenthesized value`, `alfa=bravo('charlie', 'delta')`, "", nil, `quotes must enclose the entire value (at 12)`},

		{`malformed empty key 1`, `alfa,=bravo`, "alfa", nil, `empty key (at 6)`},
		{`malformed empty key 2`, `,=alfa`, "", nil, `empty key (at 2)`},
		{`malformed empty key 3`, `''=alfa`, "", nil, `empty key (at 1)`},
		{`malformed empty key 4`, ` '' =alfa`, "", nil, `empty key (at 1)`},
		{`duplicate key last wins`, `alfa,bravo=charlie,bravo=delta`, "alfa", M{"bravo": "delta"}, ``},
		{`duplicate key first item last wins`, `foo=bar,foo=boz`, "", M{"foo": "boz"}, ``},
		{`malformed unterminated quote 1`, `alfa,'bravo=charlie`, "alfa", M{"bravo=charlie": ""}, `unterminated quote (at 6)`},
		{`malformed unterminated quote 2`, `alfa,bravo='charlie`, "alfa", M{"bravo": "charlie"}, `unterminated quote (at 12)`},
		{`malformed unterminated quote 3`, `'alfa`, "alfa", nil, `unterminated quote (at 1)`},
		{`malformed escape 1`, `a\lfa`, "alfa", nil, `invalid escape character (at 3)`},
		{`malformed escape 2`, `al\`, "al", nil, `unterminated escape sequence (at 3)`},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			tag, err := ParseWithName(test.tag)

			if test.error != "" {
				require.Error(t, err)
				assert.Equal(t, test.error, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.name, tag.Name)
				if test.opts == nil {
					assert.Empty(t, tag.Options)
				} else {
					assert.Equal(t, test.opts, tag.Options)
				}
			}
		})
	}
}

func TestParseWithName_Unquoting(t *testing.T) {
	// Test that Go struct tag quoted inputs are automatically unquoted
	tag, err := ParseWithName(`"name=value,other=key"`)
	require.NoError(t, err)
	assert.Equal(t, "", tag.Name) // No name part, just options
	assert.Equal(t, M{"name": "value", "other": "key"}, tag.Options)

	// Test that unquoted inputs work the same way
	tag2, err := ParseWithName(`name=value,other=key`)
	require.NoError(t, err)
	assert.Equal(t, "", tag2.Name)
	assert.Equal(t, M{"name": "value", "other": "key"}, tag2.Options)
}

func TestParseFuncWithName_CustomErrors(t *testing.T) {
	tests := []struct {
		name   string
		tag    string
		expErr string
		errKey string // Key that triggers error; "" for name
	}{
		{"error in name", "foo,bar=boz", "simulated error (at 1)", ""},
		{"error in key", "foo,bar=boz", "bar: simulated error (at 5)", "bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseFuncWithName(tt.tag, func(key, value string) error {
				if (tt.errKey == "" && key == "") || key == tt.errKey {
					return errSimulated
				}

				return nil
			})
			require.Error(t, err)
			assert.Equal(t, tt.expErr, err.Error())

			var errType *Error
			require.True(t, errors.As(err, &errType))
			assert.True(t, errors.Is(errType.Cause, errSimulated))
		})
	}
}

func TestParseFunc_NoEmptyKeys(t *testing.T) {
	// Regression test for bug found by fuzzing:
	// Input "0=," was calling callback with empty key
	tests := []struct {
		name  string
		input string
	}{
		{"trailing comma after key-value", "0=,"},
		{"multiple trailing commas", "key=value,,"},
		{"key-value then empty then item", "a=b,,c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseFunc(tt.input, func(key, value string) error {
				if key == "" {
					t.Errorf("received empty key in ParseFunc for input %q", tt.input)
				}

				return nil
			})
			// Empty items should be skipped, no error
			require.NoError(t, err)
		})
	}
}

func TestError_Formatting(t *testing.T) {
	// Test Error.Error() formatting for different error types
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"unterminated quote", "'unterminated", "unterminated quote"},
		{"quotes in middle", "foo'bar'", "quotes must enclose"},
		{"invalid escape", `key\nvalue`, "invalid escape"},
		{"invalid escape in key", `key\x=value`, "invalid escape"},
		{"invalid escape in option", `opt1,opt\x2`, "invalid escape"},
		{"empty key", "key1,=value", "empty key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			require.Error(t, err)

			errStr := err.Error()
			assert.Contains(t, errStr, tt.contains)

			// Verify it's a *Error type
			var parseErr *Error
			require.True(t, errors.As(err, &parseErr))
		})
	}
}

func TestHandleQuoted_EscapeAtEnd(t *testing.T) {
	// Test escape sequence at end of quoted string
	_, err := Parse(`key='value\`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unterminated escape")
}

func TestUnquoteError_DirectCall(t *testing.T) {
	// Test unquoteError.Error() method directly via Unwrap
	_, err := Parse("'incomplete")
	require.Error(t, err)

	var parseErr *Error
	require.True(t, errors.As(err, &parseErr))

	// Verify unwrapped error has Error() method
	if parseErr.Cause != nil {
		_ = parseErr.Cause.Error()
	}
}

func TestParse_KeyWithLeadingWhitespace(t *testing.T) {
	// Test key processing with leading/trailing whitespace
	tag, err := Parse("  key1  =  val1  ,  key2  =  val2  ")
	require.NoError(t, err)
	assert.Len(t, tag.Options, 2)
	assert.Equal(t, "val1", tag.Options["key1"])
	assert.Equal(t, "val2", tag.Options["key2"])
}

func TestTagTooLarge(t *testing.T) {
	// Create a tag that exceeds MaxTagLength
	largeTag := make([]byte, MaxTagLength+1)
	for i := range largeTag {
		largeTag[i] = 'a'
	}
	largeTagStr := string(largeTag)

	tests := []struct {
		name     string
		testFunc func(string) error
	}{
		{
			name: "Parse",
			testFunc: func(s string) error {
				_, err := Parse(s)

				return err
			},
		},
		{
			name: "ParseWithName",
			testFunc: func(s string) error {
				_, err := ParseWithName(s)

				return err
			},
		},
		{
			name: "ParseFunc",
			testFunc: func(s string) error {
				return ParseFunc(s, func(k, v string) error { return nil })
			},
		},
		{
			name: "ParseFuncWithName",
			testFunc: func(s string) error {
				return ParseFuncWithName(s, func(k, v string) error { return nil })
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunc(largeTagStr)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrTagTooLarge), "error should be ErrTagTooLarge")

			var parseErr *Error
			require.True(t, errors.As(err, &parseErr))
			assert.Equal(t, "tag too large", parseErr.Msg)
		})
	}
}

func TestExactlyMaxTagLength(t *testing.T) {
	// Tag exactly at the limit should work for all parse functions
	exactTag := make([]byte, MaxTagLength)
	for i := range exactTag {
		exactTag[i] = 'a'
	}
	exactTagStr := string(exactTag)

	tests := []struct {
		name     string
		testFunc func(string) error
	}{
		{
			name: "Parse",
			testFunc: func(s string) error {
				_, err := Parse(s)

				return err
			},
		},
		{
			name: "ParseWithName",
			testFunc: func(s string) error {
				_, err := ParseWithName(s)

				return err
			},
		},
		{
			name: "ParseFunc",
			testFunc: func(s string) error {
				return ParseFunc(s, func(k, v string) error { return nil })
			},
		},
		{
			name: "ParseFuncWithName",
			testFunc: func(s string) error {
				return ParseFuncWithName(s, func(k, v string) error { return nil })
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunc(exactTagStr)
			require.NoError(t, err, "exact max length should be accepted")
		})
	}
}
