# kombu

My own container runtime written in Go (WIP).

"kombu" mean kelp in Japanese.

## Features

-   Run on Linux (using namespace, cgroup...)
-   CLI tool for manage containers
-   Use OCI runtime bundle
-   Download Docker image and convert to OCI runtime bundle

## Usage

```sh
go build
sudo ./kombu download <Docker image name> <tag> # Need skopeo and umoci
sudo ./kombu create <OCI runtime bundle path>
sudo ./kombu start <container-id>
```
