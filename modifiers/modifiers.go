package modifiers

type Modifiers interface {
	Modify(string) (string, error)
	GetName() string
	Detect(string) error
	GetDetections() map[string]string
	IsApplied() bool
}
