package minikube

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/nuodb/nuodb-helm-charts/test/testlib"
	"strings"
	"testing"
	"time"
	"gotest.tools/assert"
)

func verifyAdminService(t *testing.T, namespaceName string, podName string, serviceName string, ping bool) {

	adminService := testlib.GetService(t, namespaceName, serviceName)
	assert.Equal(t, adminService.Name, serviceName)

	if ping {
		testlib.PingService(t, namespaceName, serviceName, podName)
	}
}

func verifyLBPolicy(t *testing.T, namespaceName string, podName string) {
	testlib.AwaitBalancerTerminated(t, namespaceName, "job-lb-policy")
	testlib.VerifyPolicyInstalled(t, namespaceName, podName)
}

func verifyPodKill(t *testing.T, namespaceName string, podName string, helmChartReleaseName string, nrReplicasExpected int) {
	testlib.KillAdminPod(t, namespaceName, podName)
	testlib.AwaitNrReplicasScheduled(t, namespaceName, helmChartReleaseName, nrReplicasExpected)
	testlib.AwaitPodUp(t, namespaceName, podName, 100*time.Second)
}

func verifyKillProcess(t *testing.T, namespaceName string, podName string, helmChartReleaseName string, nrReplicasExpected int) {
	testlib.KillProcess(t, namespaceName, podName)
	testlib.AwaitNrReplicasScheduled(t, namespaceName, helmChartReleaseName, nrReplicasExpected)
	testlib.AwaitPodUp(t, namespaceName, podName, 100*time.Second)
}


func verifyAllProcessesRunning(t *testing.T, namespaceName string, adminPod string, expectedNrProcesses int) {
	testlib.Await(t, func() bool {
		options := k8s.NewKubectlOptions("", "")
		options.Namespace = namespaceName

		output, err := k8s.RunKubectlAndGetOutputE(t, options, "exec", adminPod, "--", "nuocmd", "show", "domain")
		assert.NilError(t, err, "verifyAllProcessesRunning: running show domain failed")

		return strings.Count(output, "MONITORED:RUNNING") == expectedNrProcesses
	}, 30*time.Second)
}