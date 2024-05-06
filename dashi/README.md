# dashi

My own container runtime written in Go.

## Features

-   Run on Linux (using namespace, cgroup...)
-   CLI tool for manage containers
-   Use OCI runtime bundle
-   Download Docker image and convert to OCI runtime bundle

## Usage

```sh
go build
sudo ./dashi download <Docker image name> <tag> # Need skopeo and umoci
sudo ./dashi create <container-id> <OCI runtime bundle path>
sudo ./dashi start <container-id>
```
