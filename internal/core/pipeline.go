package core

 // Pipeline represents the entire CI/CD pipline
 type Pipeline struct {

 	Agent string `yaml:"agent"`              // Name of pipeline (from pipeline.yaml)
 	Stages []Stage `yaml:"stages"`		   // Ordered list of stages

   }

 // Stages represents a group of jobs
// Stages run Squentially ( stage1 --> stage2 --> stage3)

 type Stage struct {

	Name string  `yaml:"name"`               // Stage name (e.g. "build" , "test")
 	Steps []Step   `yaml:"steps"`				 // jobs inside stage
 }




