package cv

type CV struct {
	Name       string       `yaml:"name"`
	Contact    Contact      `yaml:"contact"`
	Links      []Link       `yaml:"links"`
	Summary    string       `yaml:"summary"`
	Skills     []string     `yaml:"skills"`
	Experience []Experience `yaml:"experience"`
	Projects   []Project    `yaml:"projects"`
	Education  []Education  `yaml:"education"`
	Metadata   Metadata     `yaml:"metadata"`
}

type Contact struct {
	Email    string `yaml:"email"`
	Phone    string `yaml:"phone"`
	Location string `yaml:"location"`
}

type Link struct {
	Label string `yaml:"label"`
	URL   string `yaml:"url"`
}

type Warning struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Location string `json:"location"`
}

type Experience struct {
	Company  string   `yaml:"company"`
	Title    string   `yaml:"title"`
	Location string   `yaml:"location"`
	Start    string   `yaml:"start"`
	End      string   `yaml:"end"`
	Bullets  []string `yaml:"bullets"`
}

type Project struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	URL         string   `yaml:"url"`
	Bullets     []string `yaml:"bullets"`
}

type Education struct {
	Institution string `yaml:"institution"`
	Degree      string `yaml:"degree"`
	Start       string `yaml:"start"`
	End         string `yaml:"end"`
}

type Metadata struct {
	Updated string `yaml:"updated"`
}

type Variant struct {
	Target           string   `yaml:"target"`
	MaxPages         int      `yaml:"max_pages"`
	Tone             string   `yaml:"tone"`
	SectionOrder     []string `yaml:"section_order"`
	IncludeProjects  []string `yaml:"include_projects"`
	ExcludeProjects  []string `yaml:"exclude_projects"`
	EmphasisKeywords []string `yaml:"emphasis_keywords"`
}
