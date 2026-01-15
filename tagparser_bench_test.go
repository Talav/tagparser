package tagparser

import (
	"testing"
)

// Common benchmark tag patterns.
const (
	benchTagSimple     = `json,omitempty,min=5`
	benchTagSimpleLong = `json,omitempty,required,min=5,max=100,unique,indexed`
	benchTagWithName   = `myfield,json,omitempty,min=5`
	benchTagComplex    = `name='complex,quoted=value',key=another,flag`
)

// Benchmark simple tags (fast path).
func BenchmarkParse_Simple(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(benchTagSimple)
	}
}

func BenchmarkParse_SimpleLong(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(benchTagSimpleLong)
	}
}

// Benchmark with name extraction.
func BenchmarkParseWithName_Simple(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseWithName(benchTagSimple)
	}
}

func BenchmarkParseWithName_WithName(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseWithName(benchTagWithName)
	}
}

// Benchmark complex tags with quotes.
func BenchmarkParse_Complex(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(benchTagComplex)
	}
}

func BenchmarkParse_ComplexEscapes(b *testing.B) {
	tag := `name='complex\'quoted',key=val\,ue,flag\=test`
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(tag)
	}
}

// Benchmark zero-allocation path.
func BenchmarkParseFunc_ZeroAlloc(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseFunc(benchTagSimple, func(k, v string) error { return nil })
	}
}

func BenchmarkParseFunc_WithMap(b *testing.B) {
	tag := `json,omitempty,min=5,max=100`
	opts := make(map[string]string, 4)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k := range opts {
			delete(opts, k)
		}
		_ = ParseFunc(tag, func(k, v string) error {
			opts[k] = v

			return nil
		})
	}
}

func BenchmarkParseFuncWithName_ZeroAlloc(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseFuncWithName(benchTagWithName, func(k, v string) error { return nil })
	}
}

// Benchmark real-world struct tag patterns.
func BenchmarkParse_JSONTag(b *testing.B) {
	tag := `"name,omitempty"`
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(tag)
	}
}

func BenchmarkParse_ValidateTag(b *testing.B) {
	tag := `required,email,min=8,max=100`
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(tag)
	}
}

func BenchmarkParse_DBTag(b *testing.B) {
	tag := `column:user_email,type:varchar(255),index,unique`
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(tag)
	}
}

// Benchmark worst-case scenarios.
func BenchmarkParse_ManyOptions(b *testing.B) {
	tag := `opt1,opt2,opt3,opt4,opt5,opt6,opt7,opt8,opt9,opt10,opt11,opt12,opt13,opt14,opt15`
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(tag)
	}
}

func BenchmarkParse_LongValues(b *testing.B) {
	tag := `key1='this is a very long value with lots of text',key2='another long value here',key3='and more text'`
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(tag)
	}
}
