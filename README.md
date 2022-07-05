Helm repo for the Linkerd Buoyant extension

# Buoyant Cloud setup

Head over to <https://buoyant.io/cloud> for more information about Buoyant Cloud
and the Linkerd Buoyant extension.

# Agent Install

First time repo setup:
```bash
helm repo add linkerd-buoyant https://helm.buoyant.cloud
```

## Obtain values from the Bcloud UI

To obtain a values file, head over to <https://buoyant.cloud/settings?helm=1>.
In case you have only one pair of org credentials, you will be prompted with
a dialog that contains the helm values. Otherwise, you can pick the exact
credentials pair and click on `Helm usage`. Save the values into `agent-values.yaml`

## Install
```bash
export CLUSTER_NAME=<your cluster name>
helm install --create-namespace --namespace buoyant-cloud --values agent-values.yaml --set metadata.agentName=$CLUSTER_NAME linkerd-buoyant linkerd-buoyant/linkerd-buoyant
```

# Uninstall
```bash
helm uninstall --namespace buoyant-cloud linkerd-buoyant
```

# Chart info

Helm chart releases:
<https://github.com/BuoyantIO/linkerd-buoyant/releases>

Helm repo index:
<https://github.com/BuoyantIO/linkerd-buoyant/blob/gh-pages/index.yaml>
