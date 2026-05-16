# Changelog

All notable changes to this project will be documented in this file.

## v1.2.0 - 2026-05-16

### Added

- `--json-into <file>` writes the filtered, transformed result as JSON to the
  specified file as a secondary output, independent of `--output`. Inspired by
  the [`-json-into` flag][opentofu-1.12.0] introduced in OpenTofu v1.12.0.

- `--yaml-into <file>` writes the filtered, transformed result as YAML to the
  specified file as a secondary output, independent of `--output`.

- When `--output raw` is in effect, the raw API response is written.

- Both flags support special files such as `/dev/stdout` and named pipes. Named pipes are not supported on Windows.

---

[opentofu-1.12.0]: https://opentofu.org/blog/opentofu-1-12-0/#simultaneous-human-readable-and-machine-readable-output

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).