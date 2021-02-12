# linkerd-buoyant

## Usage

```bash
$ linkerd-buoyant
This command manages the Linkerd Buoyant extension.

It enables operational control over the Buoyant Cloud Agent, providing install,
upgrade, and delete functionality

Usage:
  linkerd-buoyant [command]

Available Commands:
  help        Help about any command
  install     Output Buoyant Cloud Agent manifest for installation

Flags:
      --context string      The name of the kubeconfig context to use
  -h, --help                help for linkerd-buoyant
      --kubeconfig string   Path to the kubeconfig file to use for CLI requests (default "/home/sig/.kube/config")

Use "linkerd-buoyant [command] --help" for more information about a command.
```

## Development

Run locally:

```bash
go run main.go
```

## Release

Note the latest release:
https://github.com/BuoyantIO/linkerd-buoyant/releases

```bash
TAG=v0.0.XX
git tag $TAG
git push origin $TAG
```
