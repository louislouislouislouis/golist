package k8s

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/louislouislouislouis/repr8ducer/k8s/dkr"
	"github.com/louislouislouislouis/repr8ducer/utils"
)

func (s *K8sService) getConfigMapValue(
	ctx context.Context,
	namespace, configMapName, key string,
) (string, error) {
	configMap, err := s.Client.CoreV1().
		ConfigMaps(namespace).
		Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		utils.Log.
			Err(err).
			Msg(
				fmt.Sprintf(
					"Error during ConfigMap '%s' retrieval",
					configMapName,
				),
			)
		return "", err
	}
	value, exists := configMap.Data[key]
	if !exists {

		utils.Log.
			Err(err).
			Msg(
				fmt.Sprintf(
					"key %s not found in configMap %s",
					key,
					configMapName,
				),
			)
		return "", fmt.Errorf("key %s not found in configMap %s", key, configMapName)
	}
	return value, nil
}

func (s *K8sService) generateConfigMap(
	configMapName, namespace, volumePath string,
	ctx context.Context,
) error {
	configMap, err := s.Client.CoreV1().
		ConfigMaps(namespace).
		Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		utils.Log.
			Err(err).
			Msg(
				fmt.Sprintf(
					"Error during ConfigMap '%s' retrieval",
					configMapName,
				),
			)
		return err
	}
	for key, value := range configMap.Data {
		filePath := fmt.Sprintf("%s/%s", volumePath, key)
		writeStringFile(filePath, value)
		utils.Log.Debug().Msg(
			fmt.Sprintf(
				"File for element '%s' created: %s",
				configMapName,
				filePath,
			),
		)
	}
	return nil
}

// extractFieldRefValue est une fonction utilitaire pour récupérer les valeurs des champs FieldRef
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

func (s *K8sService) getSecretValue(
	ctx context.Context,
	namespace, secretName, key string,
) (string, error) {
	secret, err := s.Client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	value, exists := secret.Data[key]
	if !exists {
		return "", fmt.Errorf("key %s not found in secret %s", key, secretName)
	}
	return string(value), nil
}

func (s *K8sService) populateEnvContent(
	pod v1.Pod,
	container v1.Container,
	namespace string,
	ctx context.Context,
) (map[string]string, error) {
	envVars := make(map[string]string, len(container.VolumeMounts))
	for _, envVar := range container.Env {
		if envVar.Value != "" {
			envVars[envVar.Name] = envVar.Value
		}
		if envVar.ValueFrom != nil {
			switch {
			case envVar.ValueFrom.FieldRef != nil:

				fieldPath := envVar.ValueFrom.FieldRef.FieldPath
				value, err := extractFieldRefValue(pod, fieldPath)
				if err != nil {
					return nil, fmt.Errorf(
						"error extracting fieldRef value for envVar %s: %w",
						envVar.Name,
						err,
					)
				}
				envVars[envVar.Name] = value
			case envVar.ValueFrom.SecretKeyRef != nil:

				secretName := envVar.ValueFrom.SecretKeyRef.Name
				secretKey := envVar.ValueFrom.SecretKeyRef.Key

				secretValue, err := s.getSecretValue(ctx, namespace, secretName, secretKey)
				if err != nil {
					return nil, fmt.Errorf(
						"error extracting secret value for envVar %s: %w",
						envVar.Name,
						err,
					)
				}
				envVars[envVar.Name] = secretValue

			case envVar.ValueFrom.ConfigMapKeyRef != nil:

				configMapName := envVar.ValueFrom.ConfigMapKeyRef.Name
				configMapKey := envVar.ValueFrom.ConfigMapKeyRef.Key

				configMapValue, err := s.getConfigMapValue(
					ctx,
					namespace,
					configMapName,
					configMapKey,
				)
				if err != nil {
					return nil, fmt.Errorf(
						"error extracting configMap value for envVar %s: %w",
						envVar.Name,
						err,
					)
				}
				envVars[envVar.Name] = configMapValue

			default:
				utils.Log.Debug().Msg("Unknown env value ValueFrom")
			}
		}

	}

	return envVars, nil
}

func (s *K8sService) generateVolumesContent(
	pod v1.Pod,
	container v1.Container,
	namespace string,
	rootDir string,
	ctx context.Context,
) ([]string, map[string]dkr.Volume, error) {
	volumesMounts := make([]string, len(container.VolumeMounts))
	volumes := make(map[string]dkr.Volume, len(container.VolumeMounts))
	for idx, mount := range container.VolumeMounts {
		// Create the volumesMounts
		volumeDefinition := fmt.Sprintf(
			"%s/%s:%s",
			rootDir,
			mount.Name,
			mount.MountPath,
		)
		if mount.ReadOnly {
			volumeDefinition += ":ro"
		}
		volumesMounts[idx] = volumeDefinition

		// Assigning each volumesMounts a volume
		if volume, err := findVolumeByName(pod.Spec.Volumes, mount.Name); err != nil {
			return nil, nil, err
		} else {
			switch {
			case volume.ConfigMap != nil || volume.EmptyDir != nil || volume.Secret != nil || volume.Projected != nil:
				// populating the volume content
				if err := s.generateVolumeContents(*volume, namespace, rootDir, ctx); err != nil {
					return nil, nil, err
				}

				// Referencing DockerCompose Volumes
				volumes[mount.Name] = dkr.Volume{
					Driver: "local",
					DriverOpts: dkr.DriverOpts{
						Type:   "none",
						O:      "bind",
						Device: fmt.Sprintf("%s/%s", rootDir, mount.Name),
					},
				}
			default:
				utils.Log.Info().Msg(
					fmt.Sprintf("Unknown type detected for volume %s", volume.Name),
				)
			}
		}
	}
	return volumesMounts, volumes, nil
}

func (s *K8sService) generateServiceAccountTokenFile(
	namespace, volumePath string,

	ctx context.Context,
) error {
	// TODO: executer une action du type exec dans le pod
	tokenFilePath := fmt.Sprintf("%s/%s", volumePath, "service-account-token")
	tokenContent := "fake-token-content-for-testing"
	return writeStringFile(tokenFilePath, tokenContent)
}

func (s *K8sService) generateSecretFile(
	secretName, namespace, volumePath string,
	ctx context.Context,
) error {
	secret, err := s.Client.CoreV1().
		Secrets(namespace).
		Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		utils.Log.
			Err(err).
			Msg(
				fmt.Sprintf(
					"Error during Secret '%s' retrieval",
					secretName,
				),
			)
		return err
	}

	for key, value := range secret.Data {
		filePath := fmt.Sprintf("%s/%s", volumePath, key)
		writeByteFile(filePath, value)
		utils.Log.Debug().Msg(
			fmt.Sprintf(
				"File for element '%s' created: %s\n",
				secretName,
				filePath,
			),
		)
	}
	return nil
}

func (s *K8sService) generateVolumeContents(
	v v1.Volume,
	namespace string,
	rootDir string,
	ctx context.Context,
) error {
	volumePath := fmt.Sprintf("%s/%s", rootDir, v.Name)
	switch {
	case v.ConfigMap != nil:
		s.generateConfigMap(v.ConfigMap.Name, namespace, volumePath, ctx)
	case v.Secret != nil:
		s.generateSecretFile(v.Secret.SecretName, namespace, volumePath, ctx)
	case v.Projected != nil:
		for _, source := range v.Projected.Sources {
			switch {
			case source.Secret != nil:
				s.generateSecretFile(source.Secret.Name, namespace, volumePath, ctx)
			case source.ConfigMap != nil:
				s.generateConfigMap(source.ConfigMap.Name, namespace, volumePath, ctx)
			case source.DownwardAPI != nil:
				utils.Log.Info().Msg(
					fmt.Sprintf("Downward API volume detected in %s, this type is not implemented", v.Name),
				)
			case source.ServiceAccountToken != nil:
				s.generateServiceAccountTokenFile(
					namespace,
					volumePath,
					ctx,
				)
			default:
				utils.Log.Info().Msg(
					fmt.Sprintf("Unknown volumes  projected %s", v.Name),
				)
			}
		}
	case v.EmptyDir != nil:
		utils.Log.Info().Msg(
			fmt.Sprintf("TODO : generated for empty Dir %s", v.Name),
		)
	default:
		utils.Log.Info().Msg(
			fmt.Sprintf("Unknown volumes %s", v.Name),
		)
	}
	return nil
}
