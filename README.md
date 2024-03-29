# paranoidaf

![GitHub release (latest by date)](https://img.shields.io/github/v/release/echoboomer/paranoidaf)
![GitHub Release Date](https://img.shields.io/github/release-date/echoboomer/paranoidaf)
![GitHub contributors](https://img.shields.io/github/contributors/echoboomer/paranoidaf)
![GitHub issues](https://img.shields.io/github/issues/echoboomer/paranoidaf)
![GitHub license](https://img.shields.io/github/license/echoboomer/paranoidaf)
![GitHub build](https://img.shields.io/github/actions/workflow/status/echoboomer/paranoidaf/release.yaml?branch=main)

`paranoidaf` is a tool for developers who are constantly paranoid about the configuration of their Kubernetes clusters.

Configuring Kubernetes is hard. A lot of the time, we're doing it on our own with no help from anyone else. If it breaks and impacts our products, people can get upset about it.

This tool aims to help developers discover opportunities within Kubernetes clusters to make applications more resilient.

## Installation

If you have `go` installed:

```bash
go install github.com/echoboomer/paranoidaf@v0.1.1
```

Provided `GOBIN` is in `$PATH`, you should now be able to type `paranoidaf` and run the app.

Alternatively, just download the latest release from this repository and unpack it locally either to a directory in `$PATH` or somewhere you can run it directly:

```bash
tar -xvf paranoidaf_0.1.1_darwin_amd64.tar.gz -C /usr/bin
```

You should now be able to type `paranoidaf` and run the app.

## Logic

The app looks at all `Deployment` objects in the `Namespaces` (all but `kube-system`, `kube-node-lease`, `kube-public` by default, overrideable using the `--namespace` flag if you'd like to look at a specific `Namespace`) provided.

From there, details about each `Deployment` are added to a struct that keeps track of information. The `spec.selector.matchLabels` field is used to match both `HorizontalPodAutoscaler` and `PodDisruptionBudget` objects.

We make the reasonable assumption that your resources will likely share this label, usually something like `app: foobar`. If you end up with no resources returned, check these labels.

An example setup of resources is located within the `manifests/` directory for guidance.

## Usage

The app is simple and only has one command: `eval`

```bash
$ paranoidaf -h
paranoidaf helps the worried developer make sure their Kubernetes cluster is resilient.

Usage:
  paranoidaf [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  eval        Evaluate a Kubernetes cluster's configuration.
  help        Help about any command

Flags:
      --config string   config file (default is $HOME/.paranoidaf.yaml)
  -h, --help            help for paranoidaf
  -t, --toggle          Help message for toggle

Use "paranoidaf [command] --help" for more information about a command.
```

You can use `-h` to get help about the command:

```bash
$ paranoidaf eval -h
Evaluate a Kubernetes cluster's configuration.

This command looks specifically at the resiliency of your applications and
assesses their behavior during disruptive events like cluster upgrades or
Node scaling.

Usage:
  paranoidaf eval [flags]

Flags:
  -h, --help               help for eval
      --namespace string   Namespace to check. By default, all Namespaces (except for ones filtered out) are checked.

Global Flags:
      --config string   config file (default is $HOME/.paranoidaf.yaml)
```

To run the check, you can provide `eval` by itself to check all `Namespaces` except the ones filtered (shown above):

![All Namespaces](https://github.com/echoboomer/paranoidaf/blob/main/assets/sample-screenshot-1.png)

You can also provide the `--namespace` flag to check a specific `Namespace`:

![Specific Namespace](https://github.com/echoboomer/paranoidaf/blob/main/assets/sample-screenshot-2.png)

## Disclaimer

If you run into issues using the tool or find that it doesn't work for your use case(s), please feel free to open an issue and let me know about it.
