package api

import (
	"reflect"
	"testing"
	"time"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
)

func TestManageAgentStream(t *testing.T) {
	t.Run("streams and resets", func(t *testing.T) {
		fixtures := []*struct {
			testName        string
			commandsFromApi []*pb.AgentCommand
		}{
			{
				"receives commands",
				[]*pb.AgentCommand{
					createDiagnosticCommand("id1", "pod1", "ns1"),
					createDiagnosticCommand("id2", "pod2", "ns2"),
					createDiagnosticCommand("id3", "pod3", "ns3"),
					createDiagnosticCommand("id4", "pod4", "ns4"),
					createDiagnosticCommand("id5", "pod5", "ns5"),
					createDiagnosticCommand("id6", "pod6", "ns6"),
					createLogsCommand("id1", "pod1", 1),
					createLogsCommand("id2", "pod2", 2),
					createLogsCommand("id3", "pod3", 3),
					createLogsCommand("id4", "pod4", 4),
					createLogsCommand("id5", "pod5", 6),
					createLogsCommand("id6", "pod6", 5),
				},
			},
		}

		for _, tc := range fixtures {
			tc := tc
			t.Run(tc.testName, func(t *testing.T) {
				m := &MockBcloudClient{agentCommandMessages: tc.commandsFromApi}
				c := NewClient("", "", m)
				go c.Start()

				receivedCommands := []*pb.AgentCommand{}

				timeout := time.After(time.Second * 10)

			out:
				for {
					select {
					case cmd := <-c.AgentCommands():
						receivedCommands = append(receivedCommands, cmd)
						if len(receivedCommands) >= len(tc.commandsFromApi) {
							break out
						}
					case <-timeout:
						t.Fatal("test timed out")
					}
				}

				if len(receivedCommands) != len(tc.commandsFromApi) {
					t.Fatalf("Expected to receive %d commands, got: %d", len(tc.commandsFromApi), len(receivedCommands))
				}

				for i, expectedCommand := range tc.commandsFromApi {
					actualCommand := receivedCommands[i]
					if !reflect.DeepEqual(expectedCommand, actualCommand) {
						t.Fatalf("Expected command %d to be %+v, got %+v", i, expectedCommand, actualCommand)
					}
				}
			})
		}
	})
}

func createDiagnosticCommand(diagnosticID, podName string, podNamespace string) *pb.AgentCommand {
	return &pb.AgentCommand{
		Command: &pb.AgentCommand_GetProxyDiagnostics{
			GetProxyDiagnostics: &pb.GetProxyDiagnostics{
				DiagnosticId: diagnosticID,
				PodName:      podName,
				PodNamespace: podNamespace,
			},
		},
	}
}

func createLogsCommand(podName string, podNamespace string, numLines int64) *pb.AgentCommand {
	return &pb.AgentCommand{
		Command: &pb.AgentCommand_GetProxyLogs{
			GetProxyLogs: &pb.GetProxyLogs{
				PodName:      podName,
				PodNamespace: podNamespace,
				NumLines:     int32(numLines),
			},
		},
	}
}
