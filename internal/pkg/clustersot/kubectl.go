package clustersot

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sugarkube/sugarkube/internal/pkg/kapp"
	"github.com/sugarkube/sugarkube/internal/pkg/log"
	"github.com/sugarkube/sugarkube/internal/pkg/provider"
	"os/exec"
)

type KubeCtlClusterSot struct {
	ClusterSot
}

// todo - make configurable
const KUBECTL_PATH = "kubectl"
const KUBE_CONTEXT_KEY = "kube_context"

// Tests whether the cluster is online
func (c KubeCtlClusterSot) isOnline(sc *kapp.StackConfig, providerImpl provider.Provider) (bool, error) {
	providerVars := provider.GetVars(providerImpl)
	context := providerVars[KUBE_CONTEXT_KEY].(string)

	// poll `kubectl --context {{ kube_context }} get namespace`
	cmd := exec.Command(KUBECTL_PATH, "--context", context, "get", "namespace")
	err := cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			log.Debug("Cluster isn't online yet - kubectl not getting results")
			return false, nil
		}

		return false, errors.Wrap(err, "Error checking whether cluster is online")
	}

	return true, nil
}

// Tests whether all pods are Ready
func (c KubeCtlClusterSot) isReady(sc *kapp.StackConfig, providerImpl provider.Provider) (bool, error) {
	providerVars := provider.GetVars(providerImpl)
	context := providerVars[KUBE_CONTEXT_KEY].(string)

	var kubeCtlStderr, grepStdout bytes.Buffer

	kubeCtlCmd := exec.Command(KUBECTL_PATH, "--context", context, "-n", "kube-system",
		"get", "pod", "-o", "go-template=\"{{ range .items }}{{ printf \"%%s\\n\" .status.phase }}{{ end }}\"")
	kubeCtlStdout, err := kubeCtlCmd.StdoutPipe()
	kubeCtlCmd.Stderr = &kubeCtlStderr

	if err != nil {
		return false, errors.WithStack(err)
	}

	grepCmd := exec.Command("grep", "-v", "Running")
	grepCmd.Stdin = kubeCtlStdout
	grepCmd.Stdout = &grepStdout

	err = grepCmd.Start()
	if err != nil {
		return false, errors.Wrap(err, "Failed to run grep")
	}

	err = kubeCtlCmd.Start()
	if err != nil {
		return false, errors.Wrap(err, "Failed to run kubectl")
	}

	err = kubeCtlCmd.Wait()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			if kubeCtlStderr.String() != "" {
				errMsg := fmt.Sprintf("kubectl exited with %s", kubeCtlStderr.String())
				log.Fatalf(errMsg)
				return false, errors.Wrap(err, errMsg)
			} else {
				return false, nil
			}
		}

		return false, errors.Wrap(err, "kubectl terminated badly")
	}

	err = grepCmd.Wait()
	if err != nil {
		return false, errors.Wrap(err, "grep terminated badly")
	}

	// some funkiness probably with new lines means that even if grep return
	// no output, the length of its stdout buffer isn't 0, but this is
	// good enough...
	return grepStdout.Len() < 5, nil
}
