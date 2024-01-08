package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf16"
)

type TemplateJSON struct {
	Templates []Template `json:"templates"`
}

type Template struct {
	Name       string   `json:"name"`
	Filename   string   `json:"filename"`
	IconCode   string   `json:"iconCode"`
	Categories []string `json:"categories"`
	Landscape  bool     `json:"landscape,omitempty"`
}

func (t *Template) MarshalJSON() ([]byte, error) {
	var b strings.Builder
	b.WriteString("{\n")
	b.WriteString(`"name": "`)
	b.WriteString(t.Name)
	b.WriteString(`",`)

	b.WriteString(`"filename": "`)
	b.WriteString(t.Filename)
	b.WriteString(`",`)

	if t.Landscape {
		b.WriteString(`"landscape": `)
		b.WriteString("true")
		b.WriteString(`,`)
	}
	b.WriteString(`"categories": `)
	cat, _ := json.Marshal(t.Categories)
	b.Write(cat)
	b.WriteString(`,`)

	b.WriteString(`"iconCode": "`)
	codePoint := utf16.Encode([]rune(t.IconCode))
	b.WriteString(fmt.Sprintf(`\u%04x`, codePoint[0]))
	b.WriteString(`"`)

	b.WriteString("}\n")

	return []byte(b.String()), nil
}
