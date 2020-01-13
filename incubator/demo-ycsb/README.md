# YCSB Demo Helm Chart

This chart deploys the NuoDB YCSB Demo on a Kubernetes cluster using the Helm package manager.

## TL;DR;

```bash
helm install nuodb/demo-ycsb
```

## Prerequisites

- Kubernetes 1.9+
- An existing NuoDB Admin cluster has been provisioned
- An existing NuoDB demo database has been provisioned

## Installing the Chart

### Configuration

The configuration is structured where configuration values are implemented following a single-definition rule, that is, values are structured and scoped, and shared across charts; e.g. for admin, its parameters are specified once in a single values file which is used for all the charts, and the database chart can use admin values for configuring connectivity of engines to a specific admin process. The same goes for other values **shared** amongst Helm charts. A few key points here:

- values files have structure, values are scoped
- different values files for different deployments
- values files follow the single definition rule (no repeats)
- global configuration exists under its own scoped section
- each chart has its own scoped section named after it
- cloud information is used to drive availability zones (particularly)

All configurable parameters for each top-level scope is detailed below, organized by scope.

#### global.*

The purpose of this section is to specify global settings.

The following tables list the configurable parameters for the `global` option:

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `global.imageRegistry` | Global Docker image registry | `nil` |
| `global.imagePullSecrets` | Global Docker registry secret names as an array | `[]` (does not add image pull secrets to deployed pods) |

#### cloud.*

The purpose of this section is to specify the cloud provider, and specify the availability zones where a solution is deployed.

The following tables list the configurable parameters for the `cloud` option:

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `provider` | Cloud provider; permissible values include: `azure`, `amazon`, or `google` |`nil`|
| `zones` | List of availability zones to deploy to |`[]`|

For example, for the Google Cloud:

```yaml
cloud:
  provider: google
  zones:
    - us-central1-a
    - us-central1-b
    - us-central1-c
```

#### admin.*

The purpose of this section is to specify the NuoDB Admin parameters.

The following tables list the configurable parameters for the `admin` option of the admin chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `domain` | NuoDB admin cluster name | `nuodb` |
| `namespace` | Namespace where admin is deployed; when peering to an existing admin cluster provide its project name | `nuodb` |

For example, to enable an OpenShift integration, and enable routes:

```yaml
admin:
  domain: nuodb
```

#### database.*

The following tables list the configurable parameters of the database chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `name` | Database name | `demo` |

#### ycsb.*

The following tables list the configurable parameters of the YCSB chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `nameOverride` | Database name | `demo` |
| `fullnameOverride` | Number of threads | `1` |
| `loadName` | Name of the activity | `ycsb-load` |
| `workload` | YCSB workload.  Valid values are a-f. Each letter determines a different mix of read and update workload percentage generated. a= 50/50, b=95/5, c=100 read. Refer to YCSB documentation for more detail | `b` |
| `lbPolicy` | YCSB load-balancer policy. Name of an existing load-balancer policy, that has already been created using the 'nuocmd set load-balancer' command. | `ycsb-load` |
| `noOfProcesses` | Number of YCSB processes. Number of concurrent YCSB processes that will be started in each YCSB pod. Each YCSB process makes a connection to the database. | `2` |
| `noOfRows` | YCSB number of initial rows in table | `10000` |
| `noOfIterations` | YCSB number of iterations | `0` |
| `opsPerIteration` | Number of YCSB SQL operations to perform in each iteration. | `10000` |
| `maxDelay` | YCSB maximum workload delay in milliseconds (Default is 4 minutes) | `240000` |
| `dbSchema` | YCSB database schema | `USER` |
| `image.registry` | Image repository where NuoDB image is stored | `docker.io` |
| `image.repository` | Name of the Docker image | `nuodb/ycsb` |
| `image.tag` | Tag for the NuoDB Docker image | `latest` |
| `image.pullPolicy` | Image pull policy | `Always` |
| `image.pullSecrets` | Specify docker-registry secret names as an array | [] (does not add image pull secrets to deployed pods) |
| `affinity` | Affinity rules for NuoDB SM | `{}` |
| `nodeSelector` | Node selector rules for NuoDB SM | `{}` |
| `tolerations` | Tolerations for NuoDB SM | `[]` |

Verify the Helm chart:

```bash
helm install nuodb/demo-ycsb -n ycsb --debug --dry-run
```

Deploy the demo:

```bash
helm install nuodb/demo-ycsb -n ycsb
```

The command deploys NuoDB Quickstart on the Kubernetes cluster in the default configuration. The configuration section lists the parameters that can be configured during installation.

  **Tip**: List all releases using `helm list`

Wait until the deployment completes:

```bash
helm status ycsb
```

Verify the pods are running:

```bash
$ kubectl get pods
NAME                     READY   STATUS      RESTARTS   AGE
ycsb-load-xcl5f          1/1     Running     0          18h
```

  **Tip**: Wait until all processes are be in a **RUNNING** state.

Now to scale the Quickstart workload is simple enough:

```bash
$ kubectl scale rc ycsb-load --replicas=1
replicationcontroller "demo-ycsb" scaled
```

## Uninstalling the Chart

To uninstall/delete the deployment:

```bash
helm del --purge ycsb
```

The command removes all the Kubernetes components associated with the chart and deletes the release.