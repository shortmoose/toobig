# TooBig

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/shortmoose/toobig)
[![Go Report Card](https://goreportcard.com/badge/shortmoose/toobig)](https://goreportcard.com/report/shortmoose/toobig)
[![Releases](https://img.shields.io/github/release-pre/shortmoose/toobig.svg?sort=semver)](https://github.com/shortmoose/toobig/releases)
[![LICENSE](https://img.shields.io/github/license/shortmoose/toobig.svg)](https://github.com/shortmoose/toobig/blob/master/LICENSE)

TooBig converts a set of files to a pair of metadata files and files named with their SHA256. It can also convert the metadata files and SHA256 named files back to the original set of files.

This can be used for creating backups with which it is easy to verify file integrity, keep files synced across different machines, as well as various other ways this can be used.


## Install

go get github.com/shortmoose/toobig[@version]


## Docs

### toobig update <repo.yaml>

Converts a set of files into a matching set of metadata files and a hardlink to its SHA256.

### toobig restore <repo.yaml>

Converts a set of metadata and SHA256 files into the original set of files.

### toobig fsck <repo.yaml>

Verifies the data integrity of a set of files, metadata files, and SHA256s to verify everything is consistent.

The -d \<dir\> doesn't use a repo.yaml file but instead just treats the directory as a blob directory to be verified.
