package blame

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/moolen/logistis/pkg/store"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"

	"k8s.io/client-go/tools/remotecommand"
)

func execCommand(client kubernetes.Interface, config *restclient.Config, namespace, podName string,
	command string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	cmd := []string{
		"sh",
		"-c",
		command,
	}
	req := client.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(namespace).SubResource("exec")
	option := &v1.PodExecOptions{
		Command: cmd,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	if stdin == nil {
		option.Stdin = false
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return err
	}

	return nil
}

func Fetch(logger *logrus.Logger, clientset kubernetes.Interface, config *rest.Config, podLabelSelector, namespace,
	targetNamespace, targetKind, targetName string, maxHistory int) (map[string][]*store.Event, error) {
	podList, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: podLabelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list pods: %w", err)
	}
	logger.Debugf("found %d pods matching", len(podList.Items))
	if len(podList.Items) == 0 {
		return nil, fmt.Errorf("no pods found in namespaces %s with labelSelector %s", namespace, podLabelSelector)
	}
	outBuf := bytes.NewBuffer(nil)
	errBuf := bytes.NewBuffer(nil)
	cmd := fmt.Sprintf("curl -k \"https://localhost:10250/events?namespace=%s&kind=%s&name=%s&max-history=%d\"", targetNamespace, targetKind, targetName, maxHistory)
	logger.Debugf("executing command: %s", cmd)
	err = execCommand(clientset, config, namespace, podList.Items[0].ObjectMeta.Name, cmd, nil, outBuf, errBuf)
	if err != nil {
		return nil, fmt.Errorf("unable to exec command: %w / %s | %s", err, errBuf.String(), outBuf.String())
	}
	logger.Debugf("stdout: %s", outBuf.String())
	logger.Debugf("stderr: %s", errBuf.String())
	evs := make(map[string][]*store.Event)
	err = json.Unmarshal(outBuf.Bytes(), &evs)
	return evs, err
}
