package k8s

import (
	"bytes"
	"context"
	"fmt"
	"log"
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

var (
	generatedFilesBasePath   = "/tmp/repr8ducer"
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

func (s *K8sService) GetContainerFromPods(
	nms, podName string,
	ctx context.Context,
) ([]v1.Container, error) {
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
