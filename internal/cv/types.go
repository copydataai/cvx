package cv

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

type CV struct {
	Name       string       `yaml:"name" json:"name"`
	Contact    Contact      `yaml:"contact" json:"contact"`
	Links      []Link       `yaml:"links" json:"links"`
	Sources    []Source     `yaml:"sources" json:"sources,omitempty"`
	Summary    string       `yaml:"summary" json:"summary"`
	Skills     []string     `yaml:"skills" json:"skills"`
	Experience []Experience `yaml:"experience" json:"experience"`
	Projects   []Project    `yaml:"projects" json:"projects"`
	Education  []Education  `yaml:"education" json:"education"`
	Metadata   Metadata     `yaml:"metadata" json:"metadata"`
}

type Contact struct {
	Email    string `yaml:"email" json:"email"`
	Phone    string `yaml:"phone" json:"phone"`
	Location string `yaml:"location" json:"location"`
}

type Link struct {
	Label string `yaml:"label" json:"label"`
	URL   string `yaml:"url" json:"url"`
}

type Source struct {
	ID    string `yaml:"id" json:"id"`
	Type  string `yaml:"type" json:"type"`
	Label string `yaml:"label,omitempty" json:"label,omitempty"`
	URL   string `yaml:"url,omitempty" json:"url,omitempty"`
	Notes string `yaml:"notes,omitempty" json:"notes,omitempty"`
}

type Experience struct {
	Company  string   `yaml:"company" json:"company"`
	Title    string   `yaml:"title" json:"title"`
	Location string   `yaml:"location" json:"location"`
	Start    string   `yaml:"start" json:"start"`
	End      string   `yaml:"end" json:"end"`
	Bullets  []Bullet `yaml:"bullets" json:"bullets"`
}

type Project struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	URL         string   `yaml:"url" json:"url"`
	Bullets     []Bullet `yaml:"bullets" json:"bullets"`
}

type Bullet struct {
	Text     string   `yaml:"text" json:"text"`
	Source   string   `yaml:"source,omitempty" json:"source,omitempty"`
	Sources  []string `yaml:"sources,omitempty" json:"sources,omitempty"`
	Verified bool     `yaml:"verified,omitempty" json:"verified,omitempty"`
}

func (b *Bullet) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		b.Text = value.Value
		return nil
	}
	type bulletAlias Bullet
	var decoded bulletAlias
	if err := value.Decode(&decoded); err != nil {
		return err
	}
	*b = Bullet(decoded)
	return nil
}

type Education struct {
	Institution string `yaml:"institution" json:"institution"`
	Degree      string `yaml:"degree" json:"degree"`
	Start       string `yaml:"start" json:"start"`
	End         string `yaml:"end" json:"end"`
}

type Metadata struct {
	Updated string `yaml:"updated" json:"updated"`
}

type Variant struct {
	Target           string   `yaml:"target" json:"target"`
	MaxPages         int      `yaml:"max_pages" json:"max_pages"`
	Tone             string   `yaml:"tone" json:"tone"`
	SectionOrder     []string `yaml:"section_order" json:"section_order"`
	IncludeProjects  []string `yaml:"include_projects" json:"include_projects"`
	ExcludeProjects  []string `yaml:"exclude_projects" json:"exclude_projects"`
	EmphasisKeywords []string `yaml:"emphasis_keywords" json:"emphasis_keywords"`
}

type Warning struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Location string `json:"location"`
}

func (b *Bullet) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		b.Text = text
		return nil
	}
	type bulletAlias Bullet
	var decoded bulletAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*b = Bullet(decoded)
	return nil
}

func (b Bullet) String() string {
	return b.Text
}
