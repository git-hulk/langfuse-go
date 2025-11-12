package prompts

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplateCompiler_ReplacesVariables(t *testing.T) {
	compiler := newTemplateCompiler("Hello {{ name }}!")
	result := compiler.compile(map[string]any{"name": "Alice"})
	require.Equal(t, "Hello Alice!", result)
}

func TestTemplateCompiler_MissingVariablesRemain(t *testing.T) {
	compiler := newTemplateCompiler("Hello {{ name }} and {{missing}}")
	result := compiler.compile(map[string]any{"name": "Bob"})
	require.Equal(t, "Hello Bob and {{missing}}", result)
}

func TestTemplateCompiler_NilValueProducesEmptyString(t *testing.T) {
	compiler := newTemplateCompiler("{{name}}-{{other}}")
	result := compiler.compile(map[string]any{"name": nil})
	require.Equal(t, "-{{other}}", result)
}

func TestTemplateCompiler_NoVariablesProvided(t *testing.T) {
	raw := "Hello {{ name }}"
	compiler := newTemplateCompiler(raw)
	require.Equal(t, raw, compiler.compile(nil))
}

func TestTemplateCompiler_UnclosedPlaceholder(t *testing.T) {
	raw := "partial {{name"
	compiler := newTemplateCompiler(raw)
	require.Equal(t, raw, compiler.compile(map[string]any{"name": "ignored"}))
}
