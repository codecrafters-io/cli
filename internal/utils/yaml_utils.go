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
		value := strings.TrimSpace(matches[1])
		if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
			(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
			return value[1 : len(value)-1]
		}
		return value
	}
	return ""
}

func ReplaceYAMLFieldValue(content, field, newValue string) string {
	re := regexp.MustCompile(regexp.QuoteMeta(field) + `\s*:\s*([^\n\r]+)`)
	return re.ReplaceAllString(content, field+": "+newValue)
}
