package k8s

type VolumeNameNotFoundError struct {
	Message string
}

func (e *VolumeNameNotFoundError) Error() string {
	return e.Message
}
