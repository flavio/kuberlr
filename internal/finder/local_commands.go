package finder

import "strings"

// localOnlySubcommands holds the kubectl subcommands that never talk to the
// API server. The help flags are included because `kubectl -h` prints the
// usage and exits. The `__complete` entries are the hidden commands that the
// shell completion scripts invoke on every TAB.
//
//nolint:gochecknoglobals // read-only lookup table
var localOnlySubcommands = map[string]bool{
	"-h":               true,
	"--help":           true,
	"__complete":       true,
	"__completeNoDesc": true,
	"completion":       true,
	"config":           true,
	"help":             true,
	"kuberc":           true,
	"kustomize":        true,
	"options":          true,
	"plugin":           true,
}

// IsLocalOnlyCommand reports whether the given kubectl arguments describe a
// command that never needs to talk to the Kubernetes API server. For these
// commands kuberlr can skip the remote version detection and use the newest
// kubectl binary already available on the system.
//
// Only the first argument is inspected. When kubectl global flags come first,
// for example `kubectl --context foo config view`, the command is treated as
// remote. This is the safe default: kuberlr keeps its normal behaviour.
func IsLocalOnlyCommand(args []string) bool {
	if len(args) == 0 {
		// `kubectl` alone prints the usage
		return true
	}

	if localOnlySubcommands[args[0]] {
		return true
	}

	// `kubectl version` queries the server, `kubectl version --client` does not
	return args[0] == "version" && hasClientOnlyFlag(args[1:])
}

// hasClientOnlyFlag reports whether args (the arguments following the
// `version` subcommand) request client-only information, e.g.
// `--client` or `--client=true`.
func hasClientOnlyFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if arg == "--client" {
			return true
		}
		if value, ok := strings.CutPrefix(arg, "--client="); ok {
			return value == "" || value == "true"
		}
	}
	return false
}
