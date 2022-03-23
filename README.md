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

kuberlr can run on Linux, macOS and Windows.

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

You can invoke the `kuberlr` binary in a direct fashion to access its
sub-commands. For example, the `kuberlr bins` will print all the `kubectl`
binaries that are available to the user.

## How it works

kuberlr connects to the API server of your kubernetes cluster and figures
out its version.

kuberlr obtains the url of the kubernetes cluster either by looking at the
`~/.kube/config` file or by reading the contents of the file referenced by
the `KUBECONFIG` environment variable.

Once the version of the remote server is know, kuberlr looks for a compatible
kubectl binary under the `~/.kuberlr/<GOOS>-<GOARCH>/` directory and `/usr/bin`.

kuberlr reuses an already existing binary if it respects the kubectl
version skew policy, otherwise it downloads the right one from the
[upstream mirror](https://kubernetes.io/docs/tasks/tools/install-kubectl/) into
the local user cache (`~/.kuberlr/<GOOS>-<GOARCH>/`).

kuberlr names the kubectl binaries it downloads using the following naming
scheme: `kubectl<major version>.<minor version>.<patch level>`.

Finally kuberlr performs an [execve(2)](https://www.unix.com/man-page/bsd/2/EXECVE/)
syscall and leaves the control to the kubectl binary. (٭)

**Note well:** by default kuberlr will download the missing `kubectl` binaries
from the upstream mirror. This behaviour can be disabled via kuberlr's
configuration file.

The `execve` syscall is not available on Windows. On this platform another
approach is used, but the end result doesn't change. (٭)

## Reusing system-wide kubectl binaries

As pointed above kuberlr looks for a compatible kubectl binary both at user
level (`~/.kuberlr/<GOOS>-<GOARCH>/`) and at system level (`/usr/bin`).

The kubectl binaries installed at system level must respect one of these naming
schemes in order to be used:

  * `kubectl<major version>.<minor version>.<patch level>` (e.g.: `kubectl1.18.3`)
  * `kubectl<major version>.<minor version>`: this would be handled as kubectl
    version `<major version>.<minor version>.0`

## Configuration

The behaviour of kuberlr can be adjusted by creating a configuration file in
one of these locations:

  1. `/usr/etc/kuberlr.conf`: this is the location used by distributions like
    openSUSE to handle the split between `/etc` and `/usr/etc`. You can find
    more details [here](https://en.opensuse.org/openSUSE:Packaging_UsrEtc).
  1. `/etc/kuberlr.conf`
  1. `$HOME/.kuberlr/kuberlr.conf`

The configuration files are read in the order written above and merged together.
Configuration files can override the values defined by the previous ones, or
provide new ones.

The configuration file is written using the [TOML format](https://github.com/toml-lang/toml):

```toml
# Allow the download of missing kubectl binaries from kubernetes' upstream mirror
AllowDownload = true

# Directory where kubectl binaries are made accessible to all the users of the system
SystemPath = "/opt/bin"

# Timeout (sec) for requests made against the kubernetes API
Timeout = 1
```

