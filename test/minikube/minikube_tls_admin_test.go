package minikube

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nuodb/nuodb-helm-charts/test/testlib"
	"gotest.tools/assert"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
)

const ENGINE_CERTIFICATE_LOG_TEMPLATE = `Engine Certificate: Certificate #%d CN %s`

func verifySecretFields(t *testing.T, namespaceName string, secretName string, fields ...string) {
	secret := testlib.GetSecret(t, namespaceName, secretName)
	for _, field := range fields {
		_, ok := secret.Data[field]
		assert.Check(t, ok)
	}
}

func verifyKeystore(t *testing.T, namespace string, podName string, keystore string, password string, matches string) {
	options := k8s.NewKubectlOptions("", "")
	options.Namespace = namespace

	output, err := k8s.RunKubectlAndGetOutputE(t, options, "exec", podName, "--", "nuocmd", "show", "certificate", "--keystore", keystore, "--store-password", password)

	t.Log(output)
	t.Log(matches)

	assert.NilError(t, err)
	assert.Assert(t, strings.Compare(output, matches) != 0)
}

func TestKubernetesTLS(t *testing.T) {
	testlib.AwaitTillerUp(t)

	randomSuffix := strings.ToLower(random.UniqueId())

	namespaceName := fmt.Sprintf("test-admin-tls-%s", randomSuffix)
	kubectlOptions := k8s.NewKubectlOptions("", "")
	k8s.CreateNamespace(t, kubectlOptions, namespaceName)

	defer k8s.DeleteNamespace(t, kubectlOptions, namespaceName)

	defer testlib.Teardown(testlib.TEARDOWN_SECRETS)
	// create the certs...
	testlib.CreateSecret(t, namespaceName, testlib.CA_CERT_FILE, testlib.CA_CERT_SECRET, "")
	testlib.CreateSecret(t, namespaceName, testlib.NUOCMD_FILE, testlib.NUOCMD_SECRET, "")
	testlib.CreateSecretWithPassword(t, namespaceName, testlib.KEYSTORE_FILE, testlib.KEYSTORE_SECRET, testlib.SECRET_PASSWORD, "")
	testlib.CreateSecretWithPassword(t, namespaceName, testlib.TRUSTSTORE_FILE, testlib.TRUSTSTORE_SECRET, testlib.SECRET_PASSWORD, "")

	options := helm.Options{
		SetValues: map[string]string{
			"admin.replicas":               "3",
			"admin.tlsCACert.secret":       testlib.CA_CERT_SECRET,
			"admin.tlsCACert.key":          testlib.CA_CERT_FILE,
			"admin.tlsKeyStore.secret":     testlib.KEYSTORE_SECRET,
			"admin.tlsKeyStore.key":        testlib.KEYSTORE_FILE,
			"admin.tlsKeyStore.password":   testlib.SECRET_PASSWORD,
			"admin.tlsTrustStore.secret":   testlib.TRUSTSTORE_SECRET,
			"admin.tlsTrustStore.key":      testlib.TRUSTSTORE_FILE,
			"admin.tlsTrustStore.password": testlib.SECRET_PASSWORD,
			"admin.tlsClientPEM.secret":    testlib.NUOCMD_SECRET,
			"admin.tlsClientPEM.key":       testlib.NUOCMD_FILE,
		},
	}

	defer testlib.Teardown(testlib.TEARDOWN_ADMIN)

	helmChartReleaseName, namespaceName := testlib.StartAdmin(t, &options, 3, namespaceName)

	admin0 := fmt.Sprintf("%s-nuodb-0", helmChartReleaseName)

	t.Run("verifyKeystore", func(t *testing.T) {
		content, err := readAll("../../keys/default.certificate")
		assert.NilError(t, err)
		verifyKeystore(t, namespaceName, admin0, testlib.KEYSTORE_FILE, testlib.SECRET_PASSWORD, string(content))
	})

	t.Run("testDatabaseNoDirectEngineKeys", func(t *testing.T) {
		// make a copy
		localOptions := options
		localOptions.SetValues["database.sm.resources.requests.cpu"] = "500m"
		localOptions.SetValues["database.sm.resources.requests.memory"] = "1Gi"
		localOptions.SetValues["database.te.resources.requests.cpu"] = "500m"
		localOptions.SetValues["database.te.resources.requests.memory"] = "1Gi"

		defer testlib.Teardown("database")

		databaseReleaseName := testlib.StartDatabase(t, namespaceName, admin0, &localOptions)

		tePodNameTemplate := fmt.Sprintf("te-%s", databaseReleaseName)
		tePodName := testlib.GetPodName(t, namespaceName, tePodNameTemplate)
		defer testlib.GetAppLog(t, namespaceName, tePodName)

		// TE certificate is signed by the admin and the DN entry is the pod name
		// this is the 4th pod name because: #0 and #1 are trusted certs, #2 is CA, #3 is admin, #4 is engine
		expectedLogLine := fmt.Sprintf(ENGINE_CERTIFICATE_LOG_TEMPLATE, 4, tePodName)
		testlib.VerifyCertificateInLog(t, namespaceName, tePodName, expectedLogLine)
	})

	t.Run("testDatabaseDirectEngineKeys", func(t *testing.T) {
		// make a copy
		localOptions := options
		localOptions.SetValues["database.sm.resources.requests.cpu"] = "500m"
		localOptions.SetValues["database.sm.resources.requests.memory"] = "1Gi"
		localOptions.SetValues["database.te.resources.requests.cpu"] = "500m"
		localOptions.SetValues["database.te.resources.requests.memory"] = "1Gi"

		localOptions.SetValues["database.te.otherOptions.keystore"] = "/etc/nuodb/keys/nuoadmin.p12"
		localOptions.SetValues["database.sm.otherOptions.keystore"] = "/etc/nuodb/keys/nuoadmin.p12"

		defer testlib.Teardown("database")

		databaseReleaseName := testlib.StartDatabase(t, namespaceName, admin0, &localOptions)

		tePodNameTemplate := fmt.Sprintf("te-%s", databaseReleaseName)
		tePodName := testlib.GetPodName(t, namespaceName, tePodNameTemplate)
		defer testlib.GetAppLog(t, namespaceName, tePodName)

		// TE certificate is not signed by the admin and the DN entry is the generic admin name
		// this is the 3rd pod name because: #0 and #1 are trusted certs, #2 is CA, #3 is admin (and engine)
		expectedLogLine := fmt.Sprintf(ENGINE_CERTIFICATE_LOG_TEMPLATE, 3, "nuoadmin.nuodb.com")
		testlib.VerifyCertificateInLog(t, namespaceName, tePodName, expectedLogLine)
	})
}
