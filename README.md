| Go Report                                                                                                                                | Unit tests                                                                          | License |
|------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------|---------|
| [![Go Report Card](https://goreportcard.com/badge/github.com/flavio/kuberlr)](https://goreportcard.com/report/github.com/flavio/kuberlr) | ![tests](https://github.com/flavio/kuberlr/workflows/tests/badge.svg?branch=master) | [![License: Apache 2.0](https://img.shields.io/badge/License-Apache2.0-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0) |

> One kubectl to rule them all,  
> one kubectl to find them,  
> One kubectl to bring them all  
> and in the darkness bind them.  

Managing different kubernetes clusters ofter requires to keep multiple versions
of the kubectl available on the system, plus it requires some way to ensure the
right binary is used when talking with a cluster.

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

kuberlr (*kube-ruler*) is a simple wrapper for kubectl. It's main purpose is to
make it easy to manage clusters running different versions of kubernetes.

This is how kuberlr looks like:
[![asciicast](https://asciinema.org/a/326626.svg)](https://asciinema.org/a/326626)

## Usage

Put the `kuberlr` binary somewhere in your `PATH` and create a symlink named `kubectl`
pointing to it.

For example, assuming `~/bin` has a high priority inside of your `PATH`:

```
$ cp kuberlr /bin/
$ ln -s ~/bin/kuberlr ~/bin/kubectl
```

Now you can use `kubectl` as you would, behind the scene `kuberlr` will ensure
the right version of `kubectl` is used.

## How it works

kuberlr connects to the API server of your kubernetes cluster and figures
out its version.
The kubernetes cluster to talk with is obtained by looking at the `KUBECONFIG`
environment variable or at `~/.kube/config`.

Once the version of the remote server is know, kuberlr looks for the following
binary: `~/.kuberlr/<GOOS>-<GOARCH>/kubectl-<remote server major version>.<remote server minor version>-<remote server patch level>`

If the kubectl binary does not exist, kuberlr will download it from the
[upstream mirror](https://kubernetes.io/docs/tasks/tools/install-kubectl/).

Finally kuberlr performs an [execve(2)](https://www.unix.com/man-page/bsd/2/EXECVE/)
syscall and leaves the control to the kubectl binary.

## TODO

* [ ] I've tested kuberlr only on Linux. It should work also on macOS and on Windows.
  Feedback is welcome.
* [ ] Provide pre-built binaries for kuberlr (WIP)
* [ ] Test coverage, code linting,... right now this is a toy project I created
  during a weekend afternoon
* [ ] Relax the versioning constraint when downloading the kubectl version.
  Right now accessing a kubernetes 1.16.0 and a 1.16.3 cluster would result in two
  kubectl binaries being downloaded, while it would just be fine to use the latest
  one
