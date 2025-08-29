package utils

import (
	"regexp"
	"strings"
)

func ReplaceYAMLField(content, oldField, newField string) string {
	re := regexp.MustCompile(regexp.QuoteMeta(oldField) + `\s*:`)
	return re.ReplaceAllString(content, newField+":")
}

func ExtractYAMLFieldValue(content, field string) string {
	re := regexp.MustCompile(regexp.QuoteMeta(field) + `\s*:\s*([^\n\r]+)`)
	if matches := re.FindStringSubmatch(content); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func ReplaceYAMLFieldValue(content, field, newValue string) string {
	re := regexp.MustCompile(regexp.QuoteMeta(field) + `\s*:\s*([^\n\r]+)`)
	return re.ReplaceAllString(content, field+": "+newValue)
}
