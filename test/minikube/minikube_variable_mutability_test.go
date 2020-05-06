// +build long

package minikube

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/nuodb/nuodb-helm-charts/test/testlib"
	"gotest.tools/assert"
	"testing"
)

func TestKubernetesProcessOptionsMutability(t *testing.T) {
	testlib.AwaitTillerUp(t)
	defer testlib.VerifyTeardown(t)

	defer testlib.Teardown(testlib.TEARDOWN_ADMIN)

	helmChartReleaseName, namespaceName := testlib.StartAdmin(t, &helm.Options{}, 1, "")

	admin0 := fmt.Sprintf("%s-nuodb-cluster0-0", helmChartReleaseName)

	defer testlib.Teardown(testlib.TEARDOWN_DATABASE) // ensure resources allocated in called functions are released when this function exits

	options := helm.Options{
		SetValues: map[string]string{
			"database.sm.resources.requests.cpu":    testlib.MINIMAL_VIABLE_ENGINE_CPU,
			"database.sm.resources.requests.memory": testlib.MINIMAL_VIABLE_ENGINE_MEMORY,
			"database.te.resources.requests.cpu":    testlib.MINIMAL_VIABLE_ENGINE_CPU,
			"database.te.resources.requests.memory": testlib.MINIMAL_VIABLE_ENGINE_MEMORY,
		},
	}

	databaseReleaseName := testlib.StartDatabase(t, namespaceName, admin0, &options)

	t.Run("restartWithIncreasedVerbosity", func(t *testing.T) {
		assert.Assert(t, testlib.GetStringOccurenceInLog(t, namespaceName, admin0,"verbose=index") == 0)

		options.SetValues["database.sm.engineOptions.verbose"] = "index"
		options.SetValues["database.te.engineOptions.verbose"] = "index"

		testlib.RestartDatabaseWithOptions(t, namespaceName, admin0, databaseReleaseName, &options)

		verifyAllProcessesRunning(t, namespaceName, admin0, 2)

		assert.Assert(t, testlib.GetStringOccurenceInLog(t, namespaceName, admin0, "verbose=index") >= 2)
	})
}

func TestKubernetesDatabaseOptionsMutability(t *testing.T) {
	t.Skip("Database options are currently immutable")
	/*
	myArchive=0; DB=demo; hostname=sm-database-h0cz5q-nuodb-cluster0-demo-hotcopy-0
	'start sm' failed: Unable to create database: Database name=demo already exists and default engine options do not match
	*/

	testlib.AwaitTillerUp(t)
	defer testlib.VerifyTeardown(t)

	defer testlib.Teardown(testlib.TEARDOWN_ADMIN)

	helmChartReleaseName, namespaceName := testlib.StartAdmin(t, &helm.Options{}, 1, "")

	admin0 := fmt.Sprintf("%s-nuodb-cluster0-0", helmChartReleaseName)

	defer testlib.Teardown(testlib.TEARDOWN_DATABASE) // ensure resources allocated in called functions are released when this function exits

	options := helm.Options{
		SetValues: map[string]string{
			"database.sm.resources.requests.cpu":    testlib.MINIMAL_VIABLE_ENGINE_CPU,
			"database.sm.resources.requests.memory": testlib.MINIMAL_VIABLE_ENGINE_MEMORY,
			"database.te.resources.requests.cpu":    testlib.MINIMAL_VIABLE_ENGINE_CPU,
			"database.te.resources.requests.memory": testlib.MINIMAL_VIABLE_ENGINE_MEMORY,
		},
	}

	databaseReleaseName := testlib.StartDatabase(t, namespaceName, admin0, &options)

	t.Run("restartWithIncreasedVerbosity", func(t *testing.T) {
		assert.Assert(t, testlib.GetStringOccurenceInLog(t, namespaceName, admin0,"verbose=index") == 0)

		options.SetValues["database.options.verbose"] = "index"

		testlib.RestartDatabaseWithOptions(t, namespaceName, admin0, databaseReleaseName, &options)

		verifyAllProcessesRunning(t, namespaceName, admin0, 2)

		assert.Assert(t, testlib.GetStringOccurenceInLog(t, namespaceName, admin0, "verbose=index") >= 2)
	})
}
