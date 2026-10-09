package admin

import (
	"fmt"
	"strings"
)

// MaskValue replaces secret values when a config is shown in the panel.
const MaskValue = "********"

var secretKeys = map[string]bool{
	"privatekey":          true,
	"presharedkey":        true,
	"headerprotectionkey": true,
	"password":            true,
}

type lineKind int

const (
	lineOther lineKind = iota
	lineSection
	lineKeyValue
)

// scanLine classifies a config line. For key/value lines it returns the key
// (lowercased), the text up to and including the "=" and the right hand side.
func scanLine(line string) (kind lineKind, name, prefix, rhs string) {
	t := strings.TrimSpace(line)
	if t == "" || strings.HasPrefix(t, "#") || strings.HasPrefix(t, ";") {
		return lineOther, "", "", ""
	}
	if strings.HasPrefix(t, "[") {
		if end := strings.Index(t, "]"); end > 0 {
			return lineSection, strings.ToLower(strings.TrimSpace(t[1:end])), "", ""
		}
		return lineOther, "", "", ""
	}
	eq := strings.Index(line, "=")
	if eq < 0 {
		return lineOther, "", "", ""
	}
	return lineKeyValue, strings.ToLower(strings.TrimSpace(line[:eq])), line[:eq+1], strings.TrimSpace(line[eq+1:])
}

// walkSecrets calls fn for every secret key/value line, passing a stable
// identity (section, n-th section of that name, key, n-th occurrence of the
// key inside the section).
func walkSecrets(lines []string, fn func(i int, id, prefix, rhs string)) {
	section := ""
	sectionCount := map[string]int{"": 1}
	keyCount := map[string]int{}
	for i, line := range lines {
		kind, name, prefix, rhs := scanLine(line)
		switch kind {
		case lineSection:
			section = name
			sectionCount[section]++
			keyCount = map[string]int{}
		case lineKeyValue:
			if !secretKeys[name] {
				continue
			}
			keyCount[name]++
			id := fmt.Sprintf("%s#%d/%s#%d", section, sectionCount[section], name, keyCount[name])
			fn(i, id, prefix, rhs)
		}
	}
}

// MaskConfig hides the values of secret keys in a config text.
func MaskConfig(text string) string {
	lines := strings.Split(text, "\n")
	walkSecrets(lines, func(i int, _, prefix, rhs string) {
		if rhs != "" {
			lines[i] = strings.TrimRight(prefix, " ") + " " + MaskValue
		}
	})
	return strings.Join(lines, "\n")
}

// UnmaskConfig puts the real secret values from old back into every line of
// text whose value is the mask placeholder.
func UnmaskConfig(text, old string) (string, error) {
	oldValues := map[string]string{}
	walkSecrets(strings.Split(old, "\n"), func(_ int, id, _, rhs string) {
		oldValues[id] = rhs
	})

	lines := strings.Split(text, "\n")
	var firstErr error
	walkSecrets(lines, func(i int, id, prefix, rhs string) {
		if rhs != MaskValue {
			return
		}
		real, ok := oldValues[id]
		if !ok {
			if firstErr == nil {
				firstErr = fmt.Errorf("line %d: %q is masked but has no previous value; enter the real value", i+1, strings.TrimSpace(strings.TrimSuffix(prefix, "=")))
			}
			return
		}
		lines[i] = strings.TrimRight(prefix, " ") + " " + real
	})
	if firstErr != nil {
		return "", firstErr
	}
	return strings.Join(lines, "\n"), nil
}
