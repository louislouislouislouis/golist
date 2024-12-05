package k8s

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"maps"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/louislouislouislouis/repr8ducer/k8s/dkr"
	"github.com/louislouislouislouis/repr8ducer/utils"
)

var (
	generatedVolumesBasePath = "/tmp/repr8ducer"
	dockerComposeFileVersion = "3.8"
)

type K8sService struct {
	config *string
	Client *kubernetes.Clientset
}

func (s *K8sService) ListNamespace(ctx context.Context) (*v1.NamespaceList, error) {
	utils.Log.WithLevel(zerolog.DebugLevel).Msg(
		fmt.Sprintf("Start to fetch Nms"),
	)
	nms, err := s.Client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})

	utils.Log.WithLevel(zerolog.DebugLevel).Msg(
		fmt.Sprintf("Got Namespace %s", nms),
	)
	return nms, err
}

func (s *K8sService) generateDockerComposeFile(
	pod v1.Pod,
	namespace string,
	ctx context.Context,
) (string, error) {
	uuidGeneration := uuid.New()
	rootDir := fmt.Sprintf("%s/%s", generatedVolumesBasePath, uuidGeneration.String())
	dockerFile := dkr.DockerCompose{
		Version:  dockerComposeFileVersion,
		Services: make(map[string]dkr.Service, len(pod.Spec.Containers)),
		Volumes:  make(map[string]dkr.Volume, len(pod.Spec.Volumes)),
	}

	for _, container := range pod.Spec.Containers {
		volumesMounts, volumes, err := s.generateVolumesContent(
			pod,
			container,
			namespace,
			fmt.Sprintf("%s/volumes", rootDir),
			ctx,
		)
		if err != nil {
			return "", fmt.Errorf(
				"Error populating volume content of container %s : %v",
				container.Name,
				err,
			)
		}

		envVars, err := s.populateEnvContent(pod, container, namespace, ctx)
		if err != nil {
			return "", fmt.Errorf(
				"Error populating env content of container %s : %v",
				container.Name,
				err,
			)
		}

		// Merging new volumes
		maps.Copy(dockerFile.Volumes, volumes)

		// Adding Service
		dockerFile.Services[container.Name] = dkr.Service{
			Image:         container.Image,
			Volumes:       volumesMounts,
			Environnement: envVars,
		}

	}

	dockerComposePathFile := fmt.Sprintf("%s/docker-compose.yml", rootDir)

	// Generating file
	if yamlData, err := yaml.Marshal(&dockerFile); err != nil {
		return "", fmt.Errorf("Error with serializing pod %s : %v", pod.Name, err)
	} else {
		if err := writeByteFile(dockerComposePathFile, yamlData); err != nil {
			return "", fmt.Errorf("Error writing docker-compose file: %v", err)
		}
	}

	return dockerComposePathFile, nil
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

func (s *K8sService) PodToContainer(
	namespace, podName string,
	ctx context.Context,
) (string, error) {
	utils.Log.Debug().Msg(
		fmt.Sprintf("Making a pod Spec transformation to docker compose spec"),
	)
	if pod, err := s.Client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{}); err != nil {
		return "", fmt.Errorf("Error getting the pod %s : %v", podName, err)
	} else {
		pathToDockerComposeFile, err := s.generateDockerComposeFile(*pod, namespace, ctx)
		utils.Log.Debug().Msg(
			fmt.Sprintf("Pod Spec transformation to docker compose spec finished -> %s", pathToDockerComposeFile),
		)
		return fmt.Sprintf("docker compose -f %s up", pathToDockerComposeFile), err
	}
}

func (s *K8sService) ListPodsInNamespace(nms string, ctx context.Context) (*v1.PodList, error) {
	if pod, err := s.Client.CoreV1().Pods(nms).List(ctx, metav1.ListOptions{}); err != nil {
		return nil, fmt.Errorf("Error retrieving pods in nms %s : %v", nms, err)
	} else {
		return pod, nil
	}
}

func (s *K8sService) GetPod(nms, podName string, ctx context.Context) (*v1.Pod, error) {
	pod, err := s.Client.CoreV1().Pods(nms).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to get Pod %s in namespace %s : %v", podName, nms, err)
	}
	utils.Log.WithLevel(zerolog.DebugLevel).Msg(
		fmt.Sprintf("Found pod %s ", pod.Name),
	)
	return pod, nil
}

func (s *K8sService) GetContainerFromPods(nms, podName string, ctx context.Context) ([]v1.Container, error) {
	pod, err := s.Client.CoreV1().Pods(nms).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return []v1.Container{}, err
	}
	utils.Log.WithLevel(zerolog.DebugLevel).Msg(
		fmt.Sprintf("Got %d container", len(pod.Spec.Containers)),
	)
	return pod.Spec.Containers, err
}

func (s *K8sService) Exec(nms, pod, container string, ctx context.Context) (string, error) {
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

	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  nil,
		Tty:    false,
		Stdout: buf,
		Stderr: buf2,
	})
	if err != nil {
		return "", err
	}

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
