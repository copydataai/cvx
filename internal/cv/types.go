package cv

type CV struct {
	Name       string       `yaml:"name" json:"name"`
	Contact    Contact      `yaml:"contact" json:"contact"`
	Links      []Link       `yaml:"links" json:"links"`
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

type Experience struct {
	Company  string   `yaml:"company" json:"company"`
	Title    string   `yaml:"title" json:"title"`
	Location string   `yaml:"location" json:"location"`
	Start    string   `yaml:"start" json:"start"`
	End      string   `yaml:"end" json:"end"`
	Bullets  []string `yaml:"bullets" json:"bullets"`
}

type Project struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	URL         string   `yaml:"url" json:"url"`
	Bullets     []string `yaml:"bullets" json:"bullets"`
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
