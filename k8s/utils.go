package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"

	"github.com/louislouislouislouis/repr8ducer/utils"
)

func writeStringFile(filePath, value string) error {
	return writeByteFile(filePath, []byte(value))
}

func writeByteFile(filePath string, value []byte) error {
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		utils.Log.Err(err).Msgf("Error creating directories for file '%s'", filePath)
		return fmt.Errorf("error creating directories for file '%s': %v", filePath, err)
	}
	err = os.WriteFile(filePath, []byte(value), 0644)
	if err != nil {
		utils.Log.
			Err(err).
			Msg(
				fmt.Sprintf(
					"Error during creation of file '%s'",
					filePath,
				),
			)
		return fmt.Errorf(
			"Error during creation of file '%s': %v",
			filePath,
			err,
		)
	}
	return nil
}

func extractFieldRefValue(pod v1.Pod, fieldPath string) (string, error) {
	switch fieldPath {
	case "metadata.name":
		return pod.Name, nil
	case "metadata.namespace":
		return pod.Namespace, nil
	case "spec.serviceAccountName":
		return pod.Spec.ServiceAccountName, nil
	case "metadata.uid":
		return string(pod.UID), nil
	case "status.hostIP":
		if pod.Status.HostIP != "" {
			return pod.Status.HostIP, nil
		}
		return "", fmt.Errorf("fieldPath status.hostIP is not available in pod %s", pod.Name)
	case "status.podIP":
		if pod.Status.PodIP != "" {
			return pod.Status.PodIP, nil
		}
		return "", fmt.Errorf("fieldPath status.podIP is not available in pod %s", pod.Name)
	case "spec.nodeName":
		if pod.Spec.NodeName != "" {
			return pod.Spec.NodeName, nil
		}
		return "", fmt.Errorf("fieldPath spec.nodeName is not available in pod %s", pod.Name)
	default:
		utils.Log.Info().Msg(
			fmt.Sprintf("unsupported fieldPath %s", fieldPath),
		)
		return "", nil
	}
}

func findVolumeByName(volumes []v1.Volume, name string) (*v1.Volume, error) {
	for _, v := range volumes {
		if v.Name == name {
			return &v, nil
		}
	}
	return nil, &VolumeNameNotFoundError{
		Message: "Ressource non trouvée",
	}
}
