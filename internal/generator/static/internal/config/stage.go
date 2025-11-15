package config

import "fmt"

// Stage represents the deployment stage/environment
type Stage string

const (
	StageLocal      Stage = "local"
	StageProduction Stage = "production"
)

// String returns the string representation of the stage
func (s Stage) String() string {
	return string(s)
}

// IsLocal returns true if the stage is local
func (s Stage) IsLocal() bool {
	return s == StageLocal
}

// IsProduction returns true if the stage is production
func (s Stage) IsProduction() bool {
	return s == StageProduction
}

// IsValid returns true if the stage is a valid known stage
func (s Stage) IsValid() bool {
	return s == StageLocal || s == StageProduction
}

// ParseStage parses a string into a Stage enum and validates it
// Returns an error if the stage is unknown
func ParseStage(s string) (Stage, error) {
	stage := Stage(s)
	if !stage.IsValid() {
		return "", fmt.Errorf("unknown stage: %s (must be one of: %s, %s)", s, StageLocal, StageProduction)
	}
	return stage, nil
}

