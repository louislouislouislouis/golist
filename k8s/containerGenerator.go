package k8s

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/louislouislouislouis/repr8ducer/k8s/dkr"
	"github.com/louislouislouislouis/repr8ducer/modifiers"
	"github.com/louislouislouislouis/repr8ducer/utils"
)

type Generator struct {
	k8sService *K8sService
	rootDir    string
}

type GenerationResponse struct {
	Path      string
	Modifiers []modifiers.Modifiers
}

func (res GenerationResponse) GetCommand() string {
	return fmt.Sprintf("docker compose -f %s up", res.Path)
}

func NewDefaultGenerator(k8sService *K8sService) *Generator {
	return &Generator{
		rootDir:    generatedFilesBasePath,
		k8sService: k8sService,
	}
}

type containerGenerationConfig struct {
	nms               string
	pod               v1.Pod
	container         v1.Container
	generationRootDir string
	modifiers         []modifiers.Modifiers
}

func (c containerGenerationConfig) applyModifiersDetection(input string) error {
	for _, modifier := range c.modifiers {
		if err := modifier.Detect(input); err != nil {
			return fmt.Errorf(
				"Error trying to modify string %s with modifier %s : %v",
				input,
				modifier.GetName(),
				err,
			)
		}
	}
	return nil
}

func (conf containerGenerationConfig) getVolumesBaseDir() string {
	return fmt.Sprintf("%s/volumes", conf.generationRootDir)
}

func (g *Generator) generateContainerContent(
	conf containerGenerationConfig,
	ctx context.Context,
	dockerFile dkr.DockerCompose,
	initContainers []string,
) error {
	volumesMounts, volumes, err := g.generateVolumesContent(
		conf,
		ctx,
	)
	if err != nil {
		return fmt.Errorf(
			"Error populating volume content of container %s : %v",
			conf.container.Name,
			err,
		)
	}

	envVars, err := g.populateEnvContent(conf, ctx)
	if err != nil {
		return fmt.Errorf(
			"Error populating env content of container %s : %v",
			conf.container.Name,
			err,
		)
	}

	// Merging new volumes
	maps.Copy(dockerFile.Volumes, volumes)

	// Adding Service
	dockerFile.Services[conf.container.Name] = dkr.Service{
		Image:         conf.container.Image,
		Volumes:       volumesMounts,
		Environnement: envVars,
		NetworkMode:   "host",
		DependsOn:     initContainers,
		Command:       conf.container.Command,
	}
	for _, v := range conf.modifiers[0].GetDetections() {
		utils.Log.Debug().Msg(conf.modifiers[0].GetDetections()[v])
	}

	return nil
}

func (g *Generator) populateEnvContent(
	conf containerGenerationConfig,
	ctx context.Context,
) (map[string]string, error) {
	envVars := make(map[string]string, len(conf.container.VolumeMounts))
	for _, envVar := range conf.container.Env {
		if envVar.Value != "" {
			envVars[envVar.Name] = envVar.Value
		}
		if envVar.ValueFrom != nil {
			switch {
			case envVar.ValueFrom.FieldRef != nil:

				fieldPath := envVar.ValueFrom.FieldRef.FieldPath
				value, err := extractFieldRefValue(conf.pod, fieldPath)
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

				secretValue, err := g.k8sService.getSecretValue(
					ctx,
					conf.nms,
					secretName,
					secretKey,
				)
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

				configMapValue, err := g.k8sService.getConfigMapValue(
					ctx,
					conf.nms,
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

func (g *Generator) generateVolumesContent(
	conf containerGenerationConfig,
	ctx context.Context,
) ([]string, map[string]dkr.Volume, error) {
	volumesMounts := make([]string, len(conf.container.VolumeMounts))
	volumes := make(map[string]dkr.Volume, len(conf.container.VolumeMounts))
	for idx, mount := range conf.container.VolumeMounts {
		// Create the volumesMounts
		volumeDefinition := fmt.Sprintf(
			"%s/%s:%s",
			conf.getVolumesBaseDir(),
			mount.Name,
			mount.MountPath,
		)
		if mount.ReadOnly {
			volumeDefinition += ":ro"
		}
		volumesMounts[idx] = volumeDefinition

		// Assigning each volumesMounts a volume
		if volume, err := findVolumeByName(conf.pod.Spec.Volumes, mount.Name); err != nil {
			return nil, nil, err
		} else {
			switch {
			case volume.ConfigMap != nil || volume.EmptyDir != nil || volume.Secret != nil || volume.Projected != nil:
				// populating the volume content
				if err := g.generateVolumeContents(*volume, conf, ctx); err != nil {
					return nil, nil, err
				}

				// Referencing DockerCompose Volumes
				volumes[mount.Name] = dkr.Volume{
					Driver: "local",
					DriverOpts: dkr.DriverOpts{
						Type:   "none",
						O:      "bind",
						Device: fmt.Sprintf("%s/%s", conf.getVolumesBaseDir(), mount.Name),
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

func (g *Generator) generateVolumeContents(
	v v1.Volume,
	conf containerGenerationConfig,
	ctx context.Context,
) error {
	volumePath := fmt.Sprintf("%s/%s", conf.getVolumesBaseDir(), v.Name)
	switch {
	case v.ConfigMap != nil:
		g.generateConfigMap(v.ConfigMap.Name, conf, volumePath, ctx)
	case v.Secret != nil:
		g.generateSecretFile(v.Secret.SecretName, conf.nms, volumePath, ctx)
	case v.Projected != nil:
		for _, source := range v.Projected.Sources {
			switch {
			case source.Secret != nil:
				g.generateSecretFile(source.Secret.Name, conf.nms, volumePath, ctx)
			case source.ConfigMap != nil:
				g.generateConfigMap(source.ConfigMap.Name, conf, volumePath, ctx)
			case source.DownwardAPI != nil:
				utils.Log.Info().Msg(
					fmt.Sprintf("Downward API volume detected in %s, this type is not implemented", v.Name),
				)
			case source.ServiceAccountToken != nil:
				g.generateServiceAccountTokenFile(
					conf.nms,
					volumePath,
					ctx,
				)
			default:
				utils.Log.Info().Msg(
					fmt.Sprintf("Unknown volumes projected %s", v.Name),
				)
			}
		}
	case v.EmptyDir != nil:
		dir := filepath.Dir(
			fmt.Sprintf("%s/%s", volumePath, v.Name),
		)
		if err := os.MkdirAll(dir, 0777); err != nil {
			utils.Log.Err(err).Msgf("Error creating directories for emptyDir '%s'", dir)
			return fmt.Errorf("error creating directories for dir '%s': %v", dir, err)
		}
		utils.Log.Debug().Msgf("Created EmptyDir %s", dir)
	default:
		utils.Log.Info().Msg(
			fmt.Sprintf("Unknown volumes %s", v.Name),
		)
	}
	return nil
}

func (g *Generator) generateSecretFile(
	secretName, namespace, volumePath string,
	ctx context.Context,
) error {
	secret, err := g.k8sService.Client.CoreV1().
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

// TODO : remove logic for getting configMap and just accept configMap as parameter
func (g *Generator) generateConfigMap(
	configMapName string,
	conf containerGenerationConfig,
	destFileName string,
	ctx context.Context,
) error {
	configMap, err := g.k8sService.Client.CoreV1().
		ConfigMaps(conf.nms).
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
		filePath := fmt.Sprintf("%s/%s", destFileName, key)
		conf.applyModifiersDetection(value)
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

func (g *Generator) generateServiceAccountTokenFile(
	namespace, destFileName string,
	ctx context.Context,
) error {
	// TODO: executer une action du type exec dans le pod
	tokenFilePath := fmt.Sprintf("%s/%s", destFileName, "service-account-token")
	tokenContent := "fake-token-content-for-testing"
	return writeStringFile(tokenFilePath, tokenContent)
}

func (g *Generator) generateDockerComposeFile(
	pod v1.Pod,
	namespace string,
	ctx context.Context,
) (GenerationResponse, error) {
	uuidGeneration := uuid.New()
	rootDir := fmt.Sprintf("%s/%s", g.rootDir, uuidGeneration.String())
	dockerFile := dkr.DockerCompose{
		Version:  dockerComposeFileVersion,
		Services: make(map[string]dkr.Service, len(pod.Spec.Containers)),
		Volumes:  make(map[string]dkr.Volume, len(pod.Spec.Volumes)),
	}

	initContainers := make([]string, len(pod.Spec.InitContainers))
	modifiers := []modifiers.Modifiers{
		modifiers.NewUrlReplacer(),
	}
	// First generate initContainers
	for idx, container := range pod.Spec.InitContainers {
		g.generateContainerContent(containerGenerationConfig{
			pod:               pod,
			nms:               namespace,
			container:         container,
			generationRootDir: rootDir,
			modifiers:         modifiers,
		}, ctx, dockerFile, []string{})
		initContainers[idx] = container.Name
	}

	// Then generate other containers
	for _, container := range pod.Spec.Containers {
		g.generateContainerContent(containerGenerationConfig{
			pod:               pod,
			nms:               namespace,
			container:         container,
			generationRootDir: rootDir,
			modifiers:         modifiers,
		}, ctx, dockerFile, initContainers)
	}

	dockerComposePathFile := fmt.Sprintf("%s/docker-compose.yml", rootDir)

	// Generating file
	if yamlData, err := yaml.Marshal(&dockerFile); err != nil {
		return GenerationResponse{}, fmt.Errorf("Error with serializing pod %s : %v", pod.Name, err)
	} else {
		if err := writeByteFile(dockerComposePathFile, yamlData); err != nil {
			return GenerationResponse{}, fmt.Errorf("Error writing docker-compose file: %v", err)
		}
	}

	return GenerationResponse{
		Path:      dockerComposePathFile,
		Modifiers: modifiers,
	}, nil
}

// TODO: Create all necessary function by generatpr here, so the generator has no access to client
func (g *Generator) PodToContainer(
	namespace, podName string,
	ctx context.Context,
) (GenerationResponse, error) {
	utils.Log.Debug().Msg(
		fmt.Sprintf("Making a pod Spec transformation to docker compose spec"),
	)
	if pod, err := g.k8sService.GetPod(namespace, podName, ctx); err != nil {
		return GenerationResponse{}, fmt.Errorf("Error getting the pod %s : %v", podName, err)
	} else {
		response, err := g.generateDockerComposeFile(*pod, namespace, ctx)
		utils.Log.Debug().Msg(
			fmt.Sprintf("Pod Spec transformation to docker compose spec finished -> %s", response.GetCommand()),
		)
		return response, err
	}
}
