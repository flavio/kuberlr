| Go Report                                                                                                                                | Unit tests                                                                          | License |
|------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------|---------|
| [![Go Report Card](https://goreportcard.com/badge/github.com/flavio/kuberlr)](https://goreportcard.com/report/github.com/flavio/kuberlr) | [![tests](https://github.com/flavio/kuberlr/workflows/tests/badge.svg?branch=master)](https://github.com/flavio/kuberlr/actions?query=workflow%3Atests+branch%3Amaster) | [![License: Apache 2.0](https://img.shields.io/badge/License-Apache2.0-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0) |

> One kubectl to rule them all,  
> one kubectl to find them,  
> One kubectl to bring them all  
> and in the darkness bind them.  

Managing different kubernetes clusters often requires to keep multiple versions
of the kubectl available on the system, plus it poses the challenge to ensure
the right binary is used when talking with a cluster.

kubernetes defines a clear [version skew policy](https://kubernetes.io/docs/setup/release/version-skew-policy/)
for all its components. This is [what is stated about kubectl](https://kubernetes.io/docs/setup/release/version-skew-policy/#kubectl):

> kubectl is supported within one minor version (older or newer) of kube-apiserver.
>
> Example:
>
> ```
> kube-apiserver is at 1.18
> kubectl is supported at 1.19, 1.18, and 1.17
> ```

kuberlr (*kube-ruler*) is a simple wrapper for kubectl. Its main purpose is to
make it easy to manage clusters running different versions of kubernetes.

This is how kuberlr looks like in action:
[![asciicast](https://asciinema.org/a/326626.svg)](https://asciinema.org/a/326626)

## Installation

You can find pre-built binaries of kuberlr under the
[GitHub release](https://github.com/flavio/kuberlr/releases) tab.

Put the `kuberlr` binary somewhere in your `PATH` and create a symlink named `kubectl`
pointing to it.

For example, assuming `~/bin` has a high priority inside of your `PATH`:

```
$ cp kuberlr /bin/
$ ln -s ~/bin/kuberlr ~/bin/kubectl
```

## Usage

Use the `kubectl` *"fake binary"* as you usually do. Behind the scene
kuberlr will ensure a compatible version of `kubectl` is used.

## How it works

kuberlr connects to the API server of your kubernetes cluster and figures
out its version.

kuberlr obtains the url of the kubernetes cluster either by looking at the
`~/.kube/config` file or by reading the contents of the file referenced by
the `KUBECONFIG` environment variable.

Once the version of the remote server is know, kuberlr looks for a compatible
kubectl binary under the `~/.kuberlr/<GOOS>-<GOARCH>/` directory.

kuberlr reuses an already existing binary if it respects the kubectl
version skew policy, otherwise it downloads the right one from the
[upstream mirror](https://kubernetes.io/docs/tasks/tools/install-kubectl/).

kuberlr names the kubectl binaries it downloads using the following naming
scheme: `kubectl-<major version>.<minor version>.<patch level>`.

Finally kuberlr performs an [execve(2)](https://www.unix.com/man-page/bsd/2/EXECVE/)
syscall and leaves the control to the kubectl binary.
