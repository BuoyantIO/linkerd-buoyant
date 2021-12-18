Helm repo for the Linkerd Buoyant extension

# Buoyant Cloud setup

Prior to installing the Linkerd Buoyant extension, head over to
<https://buoyant.cloud>, register and create an agent. You can then download a
`values.yml` file for use with `helm install`.

# Agent Install

First time repo setup:
```bash
helm repo add linkerd-buoyant https://helm.buoyant.cloud
```

Install:
```bash
VALUES_URL=[values.yml file from buoyant.cloud]
helm install --create-namespace --namespace buoyant-cloud --values $VALUES_URL linkerd-buoyant linkerd-buoyant/linkerd-buoyant
```

Uninstall:
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
