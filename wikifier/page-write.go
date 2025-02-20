package wikifier

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

func (p *Page) Write() error {
	if !p.VarsOnly {
		return errors.New("writing pages with content is not yet supported")
	}

	var buf bytes.Buffer

	// write vars
	keys := make([]string, 0, len(p.vars))
	for k := range p.vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := p.vars[k]
		switch val := v.(type) {
		case string:
			buf.WriteString(fmt.Sprintf("@%s: %s;\n", k, escapeString(val)))
		case bool:
			if val {
				buf.WriteString(fmt.Sprintf("@%s;\n", k))
			}
		case *Map:
			writeMap(&buf, val, k)
		default:
			return errors.Errorf("page.Write: unsupported variable type for key %s", k)
		}
	}

	// write to file
	return errors.Wrap(os.WriteFile(p.Path(), buf.Bytes(), 0644), "page.Write")
}

func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "{", "\\{")
	s = strings.ReplaceAll(s, "}", "\\}")
	s = strings.ReplaceAll(s, "/*", "\\/*")
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ":", "\\:")
	return s
}

func writeMap(buf *bytes.Buffer, m *Map, prefix string) {
	keys := make([]string, 0, len(m.vars))
	for k := range m.vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := m.vars[k]
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}
		switch val := v.(type) {
		case string:
			buf.WriteString(fmt.Sprintf("@%s: %s;\n", fullKey, escapeString(val)))
		case bool:
			if val {
				buf.WriteString(fmt.Sprintf("@%s;\n", fullKey))
			}
		case *Map:
			writeMap(buf, val, fullKey)
		default:
			buf.WriteString(fmt.Sprintf("@%s: %v;\n", fullKey, val))
		}
	}
}
