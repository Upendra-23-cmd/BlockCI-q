package core

// Pipeline represents the full CI/CD pipeline
type Pipeline struct {
	Agent  string  `yaml:"agent"`
	Stages []Stage `yaml:"stages"`
}

// Stage is an ordered sequence of steps
type Stage struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`
}

// Step is a single command in a stage
type Step struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}
