package core


// job represents a unit of excution inside a stage
type Job struct {
	Name  string `yaml: "name"`  			// Job name (e.g. "complie" , "UnitTests")
	Steps  []Step `yaml: "steps"`  			// Steps inside job (usually shell/Docker commands)
}

// Step represents a single instruction inside a job

type Step struct {
	Run string 	`yaml:"run"`  				// Command to execute (e.g. "go build ./..")
}