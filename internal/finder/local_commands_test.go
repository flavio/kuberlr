package finder

import "testing"

func TestIsLocalOnlyCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{name: "no args at all", args: []string{}, expected: true},
		{name: "short help flag", args: []string{"-h"}, expected: true},
		{name: "long help flag", args: []string{"--help"}, expected: true},
		{name: "config view", args: []string{"config", "view"}, expected: true},
		{name: "config use-context", args: []string{"config", "use-context", "foo"}, expected: true},
		{name: "completion bash", args: []string{"completion", "bash"}, expected: true},
		{name: "shell completion with descriptions", args: []string{"__complete", "get", "po"}, expected: true},
		{name: "shell completion without descriptions", args: []string{"__completeNoDesc", "get", "po"}, expected: true},
		{name: "plugin list", args: []string{"plugin", "list"}, expected: true},
		{name: "kuberc view", args: []string{"kuberc", "view"}, expected: true},
		{name: "kustomize", args: []string{"kustomize", "some/dir"}, expected: true},
		{name: "options", args: []string{"options"}, expected: true},
		{name: "help get", args: []string{"help", "get"}, expected: true},
		{name: "version --client", args: []string{"version", "--client"}, expected: true},
		{name: "version --client=true with other flags", args: []string{"version", "-o", "json", "--client=true"}, expected: true},

		{name: "plain version queries the server", args: []string{"version"}, expected: false},
		{name: "version --client=false queries the server", args: []string{"version", "--client=false"}, expected: false},
		{name: "get pods", args: []string{"get", "pods"}, expected: false},
		{name: "get pods with help flag", args: []string{"get", "pods", "--help"}, expected: false},
		{name: "exec with --help meant for the child process", args: []string{"exec", "pod", "--", "ls", "--help"}, expected: false},
		{name: "explain needs server side OpenAPI data", args: []string{"explain", "pods"}, expected: false},
		{name: "api-resources queries the server", args: []string{"api-resources"}, expected: false},
		{name: "global flag before local subcommand is treated as remote", args: []string{"--context", "foo", "config", "view"}, expected: false},
		{name: "namespace flag before local subcommand is treated as remote", args: []string{"-n", "ns", "config", "view"}, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual := IsLocalOnlyCommand(tt.args)
			if actual != tt.expected {
				t.Errorf("IsLocalOnlyCommand(%v) = %v, want %v", tt.args, actual, tt.expected)
			}
		})
	}
}
