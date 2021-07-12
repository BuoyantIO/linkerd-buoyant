module github.com/buoyantio/linkerd-buoyant

go 1.16

require (
	github.com/fatih/color v1.12.0
	github.com/linkerd/linkerd2 v0.5.1-0.20210212214341-d2a40276107e
	github.com/pkg/browser v0.0.0-20201112035734-206646e67786
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	google.golang.org/grpc v1.39.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0 // indirect
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.21.2
	sigs.k8s.io/yaml v1.2.0
)
