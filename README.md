Helm repo for the Linkerd Buoyant extension

# Buoyant Cloud setup

Prior to installing the Linkerd Buoyant extension, head over to
<https://buoyant.cloud>, to register.

# Agent Install

First time repo setup:
```bash
helm repo add linkerd-buoyant https://helm.buoyant.cloud
```

## Obtain values from the Bcloud UI

To obtain a values file, head over to <https://buoyant.cloud/settings?helm=1>.
In case you have only one pair of org credentials, you will be prompted with
a dialog that contains the helm values. Otherwise, you can pick the exact
credentials pair and click on `Helm usage`. Save the values into `agen-values.yaml`

## Obtain values programatically

If you want to obtain values prograpmatically, you need to know your credentials.
If you do not know them, you can visit `https://buoyant.cloud/settings` to obtain
them. When you have the credentials execute the following set of commands:

```bash
export BUOYANT_CLOUD_CLIENT_ID=<your client id>
export BUOYANT_CLOUD_CLIENT_SECRET=<your client secret>
export CREDS=(echo -n $BUOYANT_CLOUD_CLIENT_ID:$BUOYANT_CLOUD_CLIENT_SECRET | base64 --wrap=0)
export TOKEN=(curl -s -X POST -H "Authorization: Basic $CREDS" "https://api.buoyant.cloud/token" | jq -r .access_token)
curl -s -H "Authorization: Bearer $TOKEN" "https://api.buoyant.cloud/values.yaml" > agent-values.yaml
```

## Install:
```bash
export AGENT_NAME=<your agent name>
helm install --create-namespace --namespace buoyant-cloud --values agent-values.yaml --set metadata.agentName=$AGENT_NAME linkerd-buoyant linkerd-buoyant/linkerd-buoyant
```

# Uninstall:
```bash
helm uninstall --namespace buoyant-cloud linkerd-buoyant
```

# Chart info

Helm chart `README.md` and source (unversioned):
<https://github.com/BuoyantIO/linkerd-buoyant/tree/main/charts/linkerd-buoyant>

GitHub repo:
<https://github.com/BuoyantIO/linkerd-buoyant>

Helm repo index:
<https://github.com/BuoyantIO/linkerd-buoyant/blob/gh-pages/index.yaml>
