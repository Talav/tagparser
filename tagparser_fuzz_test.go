package tagparser

import (
	"testing"
)

func FuzzParse(f *testing.F) {
	// Seed corpus with interesting inputs
	f.Add(`json,omitempty`)
	f.Add(`'quoted,value'`)
	f.Add(`key=value`)
	f.Add(`\escape\,test`)
	f.Add(`name='complex\'quoted',key=val\,ue`)
	f.Add(`foo,bar=baz`)
	f.Add(``)
	f.Add(`a`)
	f.Add(`=`)
	f.Add(`,`)
	f.Add(`'`)
	f.Add(`\`)
	f.Add(`key=`)
	f.Add(`=value`)
	f.Add(`key1,key2=val2,key3='val 3'`)
	f.Add(`"name,omitempty"`)

	f.Fuzz(func(t *testing.T, input string) {
		// Parse should never panic
		_, _ = Parse(input)
	})
}

func FuzzParseWithName(f *testing.F) {
	// Seed corpus
	f.Add(`myname,json,omitempty`)
	f.Add(`foo,bar=baz`)
	f.Add(`'quoted name',key=value`)
	f.Add(`name\,with\,commas,key=value`)
	f.Add(``)
	f.Add(`justname`)
	f.Add(`=value`)
	f.Add(`,key=value`)

	f.Fuzz(func(t *testing.T, input string) {
		// ParseWithName should never panic
		_, _ = ParseWithName(input)
	})
}

func FuzzParseFunc(f *testing.F) {
	// Seed corpus
	f.Add(`json,omitempty,min=5`)
	f.Add(`key1,key2=val2`)
	f.Add(`'quoted',escaped\,value`)

	f.Fuzz(func(t *testing.T, input string) {
		// ParseFunc should never panic
		_ = ParseFunc(input, func(k, v string) error {
			// Callback should never receive empty key in options mode
			if k == "" {
				t.Error("received empty key in ParseFunc")
			}

			return nil
		})
	})
}

func FuzzParseFuncWithName(f *testing.F) {
	// Seed corpus
	f.Add(`name,key=value`)
	f.Add(`justname`)
	f.Add(`key=value,another=val`)

	f.Fuzz(func(t *testing.T, input string) {
		// ParseFuncWithName should never panic
		_ = ParseFuncWithName(input, func(k, v string) error {
			return nil
		})
	})
}
