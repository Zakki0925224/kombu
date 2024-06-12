# yaminabe

Malware sandbox tool using [dashi](../dashi) and [nimono](../nimono).

Analyzes the log of system calls executed by the target program to see if the specified detection rules are violated.

## Detection rule

Write detection rules in TOML format. See [the sample rule file](/detection_rules/sample.toml).

## Usage

```sh
cargo run -- -t <Target program path> -d <Detection rules dir path>
```
