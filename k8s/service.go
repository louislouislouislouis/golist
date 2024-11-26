package k8s

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/louislouislouislouis/repr8ducer/utils"
)

type K8sService struct {
	config *string
	Client *kubernetes.Clientset
}

func (s *K8sService) populateVolumeContents(pod *v1.Pod, namespace string) error {
	for _, volume := range pod.Spec.Volumes {
		volumePath := fmt.Sprintf("./volumes/%s", volume.Name)

		if volume.ConfigMap != nil {
			configMap, err := s.Client.CoreV1().
				ConfigMaps(namespace).
				Get(context.TODO(), volume.ConfigMap.Name, metav1.GetOptions{})
			if err != nil {
				fmt.Printf(
					"Erreur lors de la récupération du ConfigMap '%s': %v\n",
					volume.ConfigMap.Name,
					err,
				)
				continue
			}

			// Écrire chaque entrée de données dans des fichiers séparés
			for key, value := range configMap.Data {
				filePath := fmt.Sprintf("%s/%s", volumePath, key)
				err := os.WriteFile(filePath, []byte(value), 0644)
				if err != nil {
					return fmt.Errorf(
						"Erreur lors de l'écriture du fichier '%s': %v",
						filePath,
						err,
					)
				}
				fmt.Printf(
					"Fichier pour ConfigMap '%s' créé: %s\n",
					volume.ConfigMap.Name,
					filePath,
				)
			}
		} else if volume.Secret != nil {
			secret, err := s.Client.CoreV1().
				Secrets(namespace).
				Get(context.TODO(), volume.Secret.SecretName, metav1.GetOptions{})
			if err != nil {
				fmt.Printf("Erreur lors de la récupération du Secret '%s': %v\n", volume.Secret.SecretName, err)
				continue
			}

			// Écrire chaque clé de secret dans des fichiers séparés
			for key, value := range secret.Data {
				filePath := fmt.Sprintf("%s/%s", volumePath, key)
				err := os.WriteFile(filePath, value, 0644) // Les valeurs des secrets sont []byte
				if err != nil {
					return fmt.Errorf("Erreur lors de l'écriture du fichier '%s': %v", filePath, err)
				}
				fmt.Printf("Fichier pour Secret '%s' créé: %s\n", volume.Secret.SecretName, filePath)
			}
		} else if volume.Projected != nil {
			volumePath := fmt.Sprintf("./volumes/%s", volume.Name)

			for _, source := range volume.Projected.Sources {
				if source.ConfigMap != nil {
					configMap, err := s.Client.CoreV1().
						ConfigMaps(namespace).
						Get(context.TODO(), source.ConfigMap.Name, metav1.GetOptions{})
					if err != nil {
						fmt.Printf("Erreur lors de la récupération du ConfigMap '%s' pour le volume projeté '%s': %v\n", source.ConfigMap.Name, volume.Name, err)
						continue
					}

					// Écrire les données du ConfigMap dans des fichiers
					for key, value := range configMap.Data {
						filePath := fmt.Sprintf("%s/%s", volumePath, key)
						err := os.WriteFile(filePath, []byte(value), 0644)
						if err != nil {
							fmt.Printf("Erreur lors de l'écriture du fichier '%s': %v\n", filePath, err)
							continue
						}
						fmt.Printf("Fichier pour ConfigMap '%s' (volume projeté) créé: %s\n", source.ConfigMap.Name, filePath)
					}
				}

				if source.Secret != nil {
					secret, err := s.Client.CoreV1().
						Secrets(namespace).
						Get(context.TODO(), source.Secret.Name, metav1.GetOptions{})
					if err != nil {
						fmt.Printf("Erreur lors de la récupération du Secret '%s' pour le volume projeté '%s': %v\n", source.Secret.Name, volume.Name, err)
						continue
					}

					// Écrire les données du Secret dans des fichiers
					for key, value := range secret.Data {
						filePath := fmt.Sprintf("%s/%s", volumePath, key)
						err := os.WriteFile(filePath, value, 0644) // Les valeurs des secrets sont []byte
						if err != nil {
							fmt.Printf("Erreur lors de l'écriture du fichier '%s': %v\n", filePath, err)
							continue
						}
						fmt.Printf("Fichier pour Secret '%s' (volume projeté) créé: %s\n", source.Secret.Name, filePath)
					}
				}

				if source.DownwardAPI != nil {
					// Implémentez la gestion pour DownwardAPI si nécessaire
					fmt.Printf("DownwardAPI non encore implémenté pour le volume projeté '%s'\n", volume.Name)
				}

				if source.ServiceAccountToken != nil {
					// Gérer le cas d'un ServiceAccountToken
					tokenFilePath := fmt.Sprintf("%s/%s", volumePath, "service-account-token")
					tokenContent := "fake-token-content-for-testing" // Exemple : remplacez par une récupération réelle du token si nécessaire
					err := os.WriteFile(tokenFilePath, []byte(tokenContent), 0644)
					if err != nil {
						fmt.Printf("Erreur lors de l'écriture du fichier pour ServiceAccountToken dans '%s': %v\n", tokenFilePath, err)
					} else {
						fmt.Printf("Fichier ServiceAccountToken écrit pour le volume projeté '%s': %s\n", volume.Name, tokenFilePath)
					}
				}
			}
		}
	}
	return nil
}

func (s *K8sService) ListNamespace() (*v1.NamespaceList, error) {
	utils.Log.WithLevel(zerolog.DebugLevel).Msg(
		fmt.Sprintf("Start to fetch Nms"),
	)
	nms, err := s.Client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})

	utils.Log.WithLevel(zerolog.DebugLevel).Msg(
		fmt.Sprintf("Got Namespace %s", nms),
	)
	return nms, err
}

func (s *K8sService) printProjectedSources(
	sources []v1.VolumeProjection,
	indentLevel int,
	namespace string,
) {
	indent := strings.Repeat(" ", indentLevel*2) // Ajouter des espaces pour l'indentation

	for _, source := range sources {
		if source.ConfigMap != nil {
			fmt.Printf("%s- configMap : %s\n", indent, source.ConfigMap.Name)
			configMap, err := s.Client.CoreV1().
				ConfigMaps(namespace).
				Get(context.TODO(), source.ConfigMap.Name, metav1.GetOptions{})
			if err != nil {
				fmt.Printf(
					"- Impossible de récupérer le ConfigMap '%s' : %v\n",
					source.ConfigMap.Name,
					err,
				)
				continue
			}

			// Afficher les données du ConfigMap
			fmt.Printf("- ConfigMap : %s\n", source.ConfigMap.Name)
			fmt.Println("  Données :")
			for key, value := range configMap.Data {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}
		if source.Secret != nil {
			fmt.Printf("%s- secret : %s\n", indent, source.Secret.Name)
		}
		if source.DownwardAPI != nil {
			fmt.Printf("%s- downwardAPI :\n", indent)
			// Parcourir les items du DownwardAPI
			for _, item := range source.DownwardAPI.Items {
				fmt.Printf("%s  - Path : %s\n", indent, item.Path)
				if item.FieldRef != nil {
					fmt.Printf("%s    FieldRef : %s\n", indent, item.FieldRef.FieldPath)
				}
			}
		}
		if source.ServiceAccountToken != nil {
			fmt.Printf("%s- serviceAccountToken : audience=%s, expirationSeconds=%d\n", indent,
				source.ServiceAccountToken.Audience, source.ServiceAccountToken.ExpirationSeconds)
		}
	}
}

func (s *K8sService) resolveEnvReference(
	namespace string,
	pod *v1.Pod,
	envVar v1.EnvVar,
) (string, error) {
	if envVar.ValueFrom.ConfigMapKeyRef != nil {
		// Résolution depuis un ConfigMap
		configMapRef := envVar.ValueFrom.ConfigMapKeyRef
		configMap, err := s.Client.CoreV1().
			ConfigMaps(namespace).
			Get(context.TODO(), configMapRef.Name, metav1.GetOptions{})
		if err != nil {
			return "", fmt.Errorf(
				"impossible de récupérer le ConfigMap '%s': %v",
				configMapRef.Name,
				err,
			)
		}
		value, exists := configMap.Data[configMapRef.Key]
		if !exists {
			return "", fmt.Errorf(
				"clé '%s' absente dans le ConfigMap '%s'",
				configMapRef.Key,
				configMapRef.Name,
			)
		}
		return value, nil
	} else if envVar.ValueFrom.SecretKeyRef != nil {
		// Résolution depuis un Secret
		secretRef := envVar.ValueFrom.SecretKeyRef
		secret, err := s.Client.CoreV1().Secrets(namespace).Get(context.TODO(), secretRef.Name, metav1.GetOptions{})
		if err != nil {
			return "", fmt.Errorf("impossible de récupérer le Secret '%s': %v", secretRef.Name, err)
		}
		value, exists := secret.Data[secretRef.Key]
		if !exists {
			return "", fmt.Errorf("clé '%s' absente dans le Secret '%s'", secretRef.Key, secretRef.Name)
		}
		return string(value), nil // Les secrets sont encodés en base64, donc on retourne une chaîne
	} else if envVar.ValueFrom.FieldRef != nil {
		// Résolution depuis un FieldRef
		fieldPath := envVar.ValueFrom.FieldRef.FieldPath

		// Recherche dynamique de la valeur basée sur le champ FieldPath
		switch fieldPath {
		case "metadata.name":
			return pod.ObjectMeta.Name, nil
		case "metadata.namespace":
			return pod.ObjectMeta.Namespace, nil
		case "status.hostIP":
			return pod.Status.HostIP, nil
		case "status.podIP":
			return pod.Status.PodIP, nil
		case "metadata.uid":
			return string(pod.ObjectMeta.UID), nil
		case "spec.nodeName":
			return pod.Spec.NodeName, nil
		default:
			return "", fmt.Errorf("FieldRef avec fieldPath '%s' non supporté ou introuvable", fieldPath)
		}
	}
	return "", fmt.Errorf("type de référence inconnu ou non géré")
}

func (s *K8sService) analyseVolume(volume v1.Volume, namespace string) {
	fmt.Printf("- Nom du volume : %s\n", volume.Name)

	// Identifier le type de volume
	if volume.EmptyDir != nil {
		fmt.Println("  Type : emptyDir")
	} else if volume.ConfigMap != nil {
		fmt.Printf("  Type : configMap (nom : %s)\n", volume.ConfigMap.Name)
	} else if volume.Secret != nil {
		fmt.Printf("  Type : secret (nom : %s)\n", volume.Secret.SecretName)
	} else if volume.PersistentVolumeClaim != nil {
		fmt.Printf("  Type : PVC (claim : %s)\n", volume.PersistentVolumeClaim.ClaimName)
	} else if volume.Projected != nil {
		fmt.Println("  Type : projected")
		s.printProjectedSources(volume.Projected.Sources, 2, namespace)

	} else {
		fmt.Println("  Type : autre (non spécifié)")
	}
}

func (s *K8sService) generateDockerComposeFile(pod *v1.Pod, namespace string) error {
	fileContent := []string{
		"version: '3.8'",
		"services:",
	}
	err := s.populateVolumeContents(pod, namespace)
	if err != nil {
		return fmt.Errorf("Erreur lors du peuplement des contenus des volumes: %v", err)
	}
	volumeDefinitions := make(map[string]string)

	for _, container := range pod.Spec.Containers {
		serviceName := container.Name
		fileContent = append(fileContent, fmt.Sprintf("  %s:", serviceName))
		fileContent = append(fileContent, fmt.Sprintf("    image: %s", container.Image))
		fileContent = append(fileContent, "    volumes:")

		for _, mount := range container.VolumeMounts {
			volumeName := mount.Name
			volumePath := mount.MountPath

			// Définir le mapping volume
			volumeDefinition := fmt.Sprintf("      - ./volumes/%s:%s", volumeName, volumePath)
			if mount.ReadOnly {
				volumeDefinition += ":ro"
			}
			fileContent = append(fileContent, volumeDefinition)

			// Ajouter la définition dans la section "volumes" du fichier
			volume := findVolumeByName(pod.Spec.Volumes, volumeName)
			if volume != nil {
				switch {
				case volume.EmptyDir != nil:
					// EmptyDir mappé à un répertoire local
					volumeDefinitions[volumeName] = fmt.Sprintf(
						"    %s:\n      driver: local\n      driver_opts:\n        type: none\n        o: bind\n        device: ./volumes/%s",
						volumeName,
						volumeName,
					)
				case volume.PersistentVolumeClaim != nil:
					// PVC considéré comme un volume externe
					volumeDefinitions[volumeName] = fmt.Sprintf(
						"    %s:\n      external: true",
						volumeName,
					)
				case volume.ConfigMap != nil || volume.Secret != nil || volume.Projected != nil:
					// ConfigMaps et Secrets sont mappés avec des répertoires simulés pour Docker Compose
					volumeDefinitions[volumeName] = fmt.Sprintf(
						"    %s:\n      driver: local\n      driver_opts:\n        type: none\n        o: bind\n        device: ./volumes/%s",
						volumeName,
						volumeName,
					)
				default:
					// Autres types de volumes
					volumeDefinitions[volumeName] = fmt.Sprintf(
						"    %s:\n      driver: local",
						volumeName,
					)
				}
			}
		}
	}

	// Ajouter la section volumes
	if len(volumeDefinitions) > 0 {
		fileContent = append(fileContent, "volumes:")
		for _, volumeDef := range volumeDefinitions {
			fileContent = append(fileContent, volumeDef)
		}
	}

	// Écriture dans un fichier docker-compose.yml
	fileName := "docker-compose.yml"
	err = os.WriteFile(fileName, []byte(strings.Join(fileContent, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("Error writing docker-compose file: %v", err)
	}

	fmt.Printf("Fichier Docker Compose généré : %s\n", fileName)

	// Création des répertoires locaux pour les volumes
	err = createLocalVolumeDirs(volumeDefinitions)
	if err != nil {
		return fmt.Errorf("Error creating local directories for volumes: %v", err)
	}

	return nil
}

func createLocalVolumeDirs(volumeDefinitions map[string]string) error {
	for volumeName := range volumeDefinitions {
		volumePath := fmt.Sprintf("./volumes/%s", volumeName)
		err := os.MkdirAll(volumePath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create volume directory %s: %v", volumePath, err)
		}
		fmt.Printf("Répertoire pour le volume '%s' créé : %s\n", volumeName, volumePath)
	}
	return nil
}

func findVolumeByName(volumes []v1.Volume, name string) *v1.Volume {
	for _, v := range volumes {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

func (s *K8sService) listVolumes(namespace, podName string, ctx context.Context) error {
	pod, err := s.Client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Error getting the pod %s : %v", podName, err)
	}
	volumes := make(map[string]v1.Volume)
	for _, volume := range pod.Spec.Volumes {

		volumes[volume.Name] = volume
		fmt.Printf("- Nom du volume : %s\n", volume.Name)
		switch {
		case volume.EmptyDir != nil:
			fmt.Println("  Type : emptyDir")
		case volume.ConfigMap != nil:
			fmt.Printf("  Type : configMap (nom : %s)\n", volume.ConfigMap.Name)
		case volume.Secret != nil:
			fmt.Printf("  Type : secret (nom : %s)\n", volume.Secret.SecretName)
		case volume.PersistentVolumeClaim != nil:
			fmt.Printf("  Type : PVC (claim : %s)\n", volume.PersistentVolumeClaim.ClaimName)
		case volume.Projected != nil:
			fmt.Println("  Type : projected")
			s.printProjectedSources(volume.Projected.Sources, 2, namespace)
		default:
			fmt.Println("  Type : autre (non spécifié)")
		}
	}

	fmt.Printf("\nMontages des volumes dans les conteneurs du Pod '%s':\n", podName)
	for _, container := range pod.Spec.Containers {
		fmt.Printf("Conteneur : %s\n", container.Name)
		for _, mount := range container.VolumeMounts {
			fmt.Printf("  Monté : %s -> %s\n", mount.Name, mount.MountPath)
			if mount.ReadOnly {
				fmt.Println("    Lecture seule : Oui")
			} else {
				fmt.Println("    Lecture seule : Non")
			}
		}
	}
	s.generateDockerComposeFile(pod, namespace)

	return nil
}

func (s *K8sService) GetEnvFromContainer(
	nms, podName, container string,
	ctx context.Context,
) error {
	// Récupérer les informations du pod
	pod, err := s.Client.CoreV1().Pods(nms).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("impossible de récupérer le pod : %v", err)
	}

	// Parcourir les conteneurs et afficher les variables d'environnement
	for _, container := range pod.Spec.Containers {
		fmt.Printf("Container: %s\n", container.Name)

		if len(container.Env) == 0 {
			fmt.Println("  Pas de variables d'environnement définies")
			continue
		}

		for _, envVar := range container.Env {
			if envVar.Value != "" {
				fmt.Printf("  %s: %s\n", envVar.Name, envVar.Value)
			} else if envVar.ValueFrom != nil {
				// Résoudre les références
				value, err := s.resolveEnvReference(nms, pod, envVar)
				if err != nil {
					fmt.Printf("  %s: erreur lors de la résolution (%v)\n", envVar.Name, err)
				} else {
					fmt.Printf("  %s: %s\n", envVar.Name, value)
				}
			}
		}
	}
	return nil
}

func (s *K8sService) ListPodsInNamespace(nms string) (*v1.PodList, error) {
	pod, err := s.Client.CoreV1().Pods(nms).List(context.TODO(), metav1.ListOptions{})

	// Todo Handle error
	utils.Log.WithLevel(zerolog.DebugLevel).Msg(
		fmt.Sprintf("Got pods %s", pod),
	)

	return pod, err
}

func (s *K8sService) GetPod(nms, podName string) (*v1.Pod, error) {
	pod, err := s.Client.CoreV1().Pods(nms).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		// Todo Handle error and context
		panic(err.Error())
	}
	utils.Log.WithLevel(zerolog.DebugLevel).Msg(
		fmt.Sprintf("Got pods %s", pod),
	)
	return pod, err
}

func (s *K8sService) GetContainerFromPods(nms, podName string) ([]v1.Container, error) {
	pod, err := s.Client.CoreV1().Pods(nms).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return []v1.Container{}, err
	}
	utils.Log.WithLevel(zerolog.DebugLevel).Msg(
		fmt.Sprintf("Got %d container", len(pod.Spec.Containers)),
	)
	return pod.Spec.Containers, err
}

func (s *K8sService) Exec(nms, pod, container string) (string, error) {
	req := s.Client.CoreV1().
		RESTClient().
		Post().
		Resource("pods").
		Name(pod).
		Namespace(nms).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Container: container,
		Command:   []string{"cat", "config/application.yaml"},
		Stdin:     true,
		Stdout:    true,
		Stderr:    false,
	}, scheme.ParameterCodec)
	// find / -type f -name application.yaml 2>/dev/null | xargs cat

	kubeCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restCfg, err := kubeCfg.ClientConfig()

	exec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", req.URL())
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}

	utils.Log.Debug().Msg("Heeyyy")

	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdin:  nil,
		Tty:    false,
		Stdout: buf,
		Stderr: buf2,
	})
	if err != nil {
		return "", err
	}

	utils.Log.Debug().Msg("Heeyyy")
	test3 := buf.String()

	// Analyse du YAML
	var data map[string]interface{}
	err = yaml.Unmarshal([]byte(test3), &data)
	if err != nil {
		log.Fatalf("Erreur lors de l'analyse YAML : %v", err)
	}

	// Transformation en slice de string
	flattened := flattenYaml("", data)

	// Impression du résultat
	// fmt.Println(strings.Join(flattened, "\n"))

	return strings.Join(flattened, "\n"), nil
}

// Fonction pour transformer une structure en slice de string
func flattenYaml(prefix string, data map[string]interface{}) []string {
	var result []string
	for key, value := range data {
		fullKey := prefix + key

		// Si la valeur est une autre map, on appelle récursivement flattenYaml
		if reflect.TypeOf(value).Kind() == reflect.Map {
			nestedMap := value.(map[string]interface{})
			result = append(result, flattenYaml(fullKey+".", nestedMap)...)
		} else {
			// Sinon, on ajoute le paramètre dans le format demandé
			formattedValue := fmt.Sprintf("--env \"%s='%v'\"", fullKey, value)
			result = append(result, formattedValue)
		}
	}
	return result
}
