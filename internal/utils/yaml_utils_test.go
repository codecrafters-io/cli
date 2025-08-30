package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceYAMLField(t *testing.T) {
	content := `# Set this to true if you want debug logs.
debug: false

# Use this to change the Go version used to run your code
language_pack: go-1.19
`
	result := ReplaceYAMLField(content, "language_pack", "buildpack")
	expected := `# Set this to true if you want debug logs.
debug: false

# Use this to change the Go version used to run your code
buildpack: go-1.19
`
	assert.Equal(t, expected, result)
}

func TestReplaceYAMLFieldWithDoubleQuotes(t *testing.T) {
	content := `# Configuration file
debug: "false"
language_pack: "go-1.19"
`
	result := ReplaceYAMLField(content, "language_pack", "buildpack")
	expected := `# Configuration file
debug: "false"
buildpack: "go-1.19"
`
	assert.Equal(t, expected, result)
}

func TestReplaceYAMLFieldWithSingleQuotes(t *testing.T) {
	content := `# CodeCrafters config
debug: 'false'
language_pack: 'go-1.19'
`
	result := ReplaceYAMLField(content, "language_pack", "buildpack")
	expected := `# CodeCrafters config
debug: 'false'
buildpack: 'go-1.19'
`
	assert.Equal(t, expected, result)
}

func TestReplaceYAMLFieldWithWeirdWhitespace(t *testing.T) {
	content := `# Comments preserved
debug:   false   
language_pack	:	go-1.19	
# End comment
`
	result := ReplaceYAMLField(content, "language_pack", "buildpack")
	expected := `# Comments preserved
debug:   false   
buildpack:	go-1.19	
# End comment
`
	assert.Equal(t, expected, result)
}

func TestExtractYAMLFieldValue(t *testing.T) {
	content := `# Available versions: rust-1.88
debug: false
buildpack: rust-1.88
`
	value := ExtractYAMLFieldValue(content, "buildpack")
	assert.Equal(t, "rust-1.88", value)

	value = ExtractYAMLFieldValue(content, "nonexistent")
	assert.Equal(t, "", value)
}

func TestExtractYAMLFieldValueWithDoubleQuotes(t *testing.T) {
	content := `# Configuration
debug: "false"
buildpack: "rust-1.88"
`
	value := ExtractYAMLFieldValue(content, "buildpack")
	assert.Equal(t, "rust-1.88", value)
}

func TestExtractYAMLFieldValueWithSingleQuotes(t *testing.T) {
	content := `# Project settings
debug: 'false'
buildpack: 'rust-1.88'
`
	value := ExtractYAMLFieldValue(content, "buildpack")
	assert.Equal(t, "rust-1.88", value)
}

func TestExtractYAMLFieldValueWithWeirdWhitespace(t *testing.T) {
	content := `# Messy formatting but comments preserved
debug:   false   
buildpack	:	rust-1.88	
# Another comment
`
	value := ExtractYAMLFieldValue(content, "buildpack")
	assert.Equal(t, "rust-1.88", value)
}

func TestReplaceYAMLFieldValue(t *testing.T) {
	content := `# Set this to true if you want debug logs.
debug: false

# Use this to change the Rust version used to run your code
# Available versions: rust-1.88
buildpack: rust-1.88
`
	result := ReplaceYAMLFieldValue(content, "buildpack", "rust-1.89")
	expected := `# Set this to true if you want debug logs.
debug: false

# Use this to change the Rust version used to run your code
# Available versions: rust-1.88
buildpack: rust-1.89
`
	assert.Equal(t, expected, result)
}

func TestReplaceYAMLFieldValueWithDoubleQuotes(t *testing.T) {
	content := `# Config with quotes
debug: "false"
buildpack: "rust-1.88"
`
	result := ReplaceYAMLFieldValue(content, "buildpack", "rust-1.89")
	expected := `# Config with quotes
debug: "false"
buildpack: rust-1.89
`
	assert.Equal(t, expected, result)
}

func TestReplaceYAMLFieldValueWithSingleQuotes(t *testing.T) {
	content := `# Single quoted values
debug: 'false'
buildpack: 'rust-1.88'
# End of config
`
	result := ReplaceYAMLFieldValue(content, "buildpack", "rust-1.89")
	expected := `# Single quoted values
debug: 'false'
buildpack: rust-1.89
# End of config
`
	assert.Equal(t, expected, result)
}

func TestReplaceYAMLFieldValueWithWeirdWhitespace(t *testing.T) {
	content := `# Weird spacing but comments intact
debug:   false   
buildpack	:	rust-1.88	
# Final comment
`
	result := ReplaceYAMLFieldValue(content, "buildpack", "rust-1.89")
	expected := `# Weird spacing but comments intact
debug:   false   
buildpack: rust-1.89
# Final comment
`
	assert.Equal(t, expected, result)
}
