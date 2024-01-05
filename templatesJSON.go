package main

type TemplateJSON struct {
	Templates []struct {
		Name       string   `json:"name"`
		Filename   string   `json:"filename"`
		IconCode   string   `json:"iconCode"`
		Categories []string `json:"categories"`
		Landscape  bool     `json:"landscape,omitempty"`
	} `json:"templates"`
}
