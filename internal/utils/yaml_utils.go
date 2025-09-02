package utils

import (
	"regexp"
)

func ReplaceYAMLField(content, oldField, newField string) string {
	re := regexp.MustCompile(regexp.QuoteMeta(oldField) + `\s*:`)
	return re.ReplaceAllString(content, newField+":")
}

func ReplaceYAMLFieldValue(content, field, newValue string) string {
	re := regexp.MustCompile(regexp.QuoteMeta(field) + `\s*:\s*([^\n\r]+)`)
	return re.ReplaceAllString(content, field+": "+newValue)
}
