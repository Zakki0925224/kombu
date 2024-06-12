# kombu

My own container runtime project.

## [dashi](/dashi)

My own container runtime written in Go.

## [nimono](/nimono)

Linux system call logger using eBPF written in Go.

## [yaminabe](/yaminabe)

Malware sandbox tool written in Rust using [dashi](/dashi) and [nimono](/nimono).

## Requipments

-   Linux kernel 5.5?~
-   Rust
-   Go
-   Bpftool
-   Clang

## Usage

### Build and run malware sandbox

```sh
python3 ./task.py task_run
```
