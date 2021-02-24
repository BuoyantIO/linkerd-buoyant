# linkerd-buoyant

[![Actions](https://github.com/BuoyantIO/linkerd-buoyant/actions/workflows/actions.yml/badge.svg)](https://github.com/BuoyantIO/linkerd-buoyant/actions/workflows/actions.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/buoyantio/linkerd-buoyant)](https://goreportcard.com/report/github.com/buoyantio/linkerd-buoyant)
[![GitHub license](https://img.shields.io/github/license/buoyantio/linkerd-buoyant.svg)](LICENSE)

The Linkerd Buoyant extension is a CLI tool for managing the Buoyant Cloud
Agent.

## Install

To install the CLI, run:

```bash
curl https://buoyant.cloud/install | sh
```

Alternatively, you can download the CLI directly via the
[releases page](https://github.com/BuoyantIO/linkerd-buoyant/releases).

## Usage

```bash
$ linkerd-buoyant
linkerd-buoyant manages the Buoyant Cloud Agent.

It enables operational control over the Buoyant Cloud Agent, providing install,
upgrade, and delete functionality

Usage:
  linkerd-buoyant [command]

Available Commands:
  check       Check the Buoyant Cloud Agent installation for potential problems
  help        Help about any command
  install     Output Buoyant Cloud Agent manifest for installation
  uninstall   Output Kubernetes resources to uninstall the Buoyant Cloud Agent
  version     Print the CLI and Agent version information

Flags:
      --context string      The name of the kubeconfig context to use
  -h, --help                help for linkerd-buoyant
      --kubeconfig string   Path to the kubeconfig file to use for CLI requests (default "/home/sig/.kube/config")
  -v, --verbose             Turn on debug logging

Use "linkerd-buoyant [command] --help" for more information about a command.
```

## Development

Run locally:
```bash
go run main.go
```

Run with a version number:
```bash
go run -ldflags "-s -w -X github.com/buoyantio/linkerd-buoyant/pkg/version.Version=vX.Y.Z" main.go version
```

Test against a local server:
```bash
go run main.go --bcloud-server http://localhost:8084
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
