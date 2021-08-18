# linkerd-buoyant

[![Actions](https://github.com/BuoyantIO/linkerd-buoyant/actions/workflows/actions.yml/badge.svg)](https://github.com/BuoyantIO/linkerd-buoyant/actions/workflows/actions.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/buoyantio/linkerd-buoyant)](https://goreportcard.com/report/github.com/buoyantio/linkerd-buoyant)
[![GitHub license](https://img.shields.io/github/license/buoyantio/linkerd-buoyant.svg)](LICENSE)

The Linkerd Buoyant extension connects your
[Linkerd](https://linkerd.io)-enabled Kubernetes cluster to
[Buoyant Cloud](https://buoyant.io/cloud), the global platform health dashboard for
Linkerd.

This repo consists of two components:
- [`agent`](agent): Runs on your Kubernetes cluster.
- [`cli`](cli): Runs locally or wherever you install the Linkerd CLI.

## Install CLI

To install the CLI, run:

```bash
curl -sL https://buoyant.cloud/install | sh
```

Alternatively, you can download the CLI directly via the
[releases page](https://github.com/BuoyantIO/linkerd-buoyant/releases).

### Usage

```bash
$ linkerd-buoyant
linkerd-buoyant manages the Buoyant Cloud agent.

It enables operational control over the Buoyant Cloud agent, providing install,
upgrade, and delete functionality.

Usage:
  linkerd-buoyant [command]

Available Commands:
  check       Check the Buoyant Cloud agent installation for potential problems
  help        Help about any command
  install     Output Buoyant Cloud agent manifest for installation
  uninstall   Output Kubernetes manifest to uninstall the Buoyant Cloud agent
  version     Print the CLI and agent version information

Flags:
      --context string      The name of the kubeconfig context to use
  -h, --help                help for linkerd-buoyant
      --kubeconfig string   Path to the kubeconfig file to use for CLI requests (default "/home/sig/.kube/config")
  -v, --verbose             Turn on debug logging

Use "linkerd-buoyant [command] --help" for more information about a command.
```

## Development

### Agent

Build and run:
```bash
bin/go-run agent
```

Docker build:
```bash
docker buildx build -f agent/Dockerfile -t ghcr.io/buoyantio/linkerd-buoyant:latest .
```

### CLI

Build and run:
```bash
bin/go-run cli
```

Run with a version number:
```bash
go run -ldflags "-s -w -X github.com/buoyantio/linkerd-buoyant/cli/pkg/version.Version=vX.Y.Z" cli/main.go version
```

Test against a local server:
```bash
bin/go-run cli --bcloud-server http://localhost:8084 check
```

### Protobuf

The generated protobuf bindings in `gen` come from the `proto` directory in this
repo. If you make changes there, re-generate them with:

```bash
bin/gen-proto
```

### Testing

```bash
go test -race -cover ./...
bin/lint
```

## Release

Note the latest release:
https://github.com/BuoyantIO/linkerd-buoyant/releases

```bash
TAG=v0.0.XX
git tag $TAG
git push origin $TAG
```

## License

Copyright 2021 Buoyant, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
these files except in compliance with the License. You may obtain a copy of the
License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
