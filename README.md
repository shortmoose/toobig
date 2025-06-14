# TooBig

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/shortmoose/toobig)
[![Go Report Card](https://goreportcard.com/badge/shortmoose/toobig)](https://goreportcard.com/report/shortmoose/toobig)
[![Releases](https://img.shields.io/github/release-pre/shortmoose/toobig.svg?sort=semver)](https://github.com/shortmoose/toobig/releases)
[![LICENSE](https://img.shields.io/github/license/shortmoose/toobig.svg)](https://github.com/shortmoose/toobig/blob/master/LICENSE)

TooBig was created from a need to manage large files like photos, videos, and other files.
Over the last number of years I have found it has worked wonderfully
for managing my decades of family photos, but also for managing photo and video
assets for development projects, and other uses here and there.

TooBig converts a set of files
into a set of refs and blobs. A blob is the original
file, but the name has been changed to the 64 character SHA-256 checksum of the
contents of the file. A ref is a small file that has the same name and directory
structure of the original file, but the file itself is tiny, containing just the
SHA-256 checksum of the original file and a timestamp.

`toobig update` converts files to refs/blobs.
`toobig restore` converts refs/blobs to files.

## Why Use TooBig?

TooBig's architecture of separating file metadata (refs) from file content (blobs) provides several powerful advantages:

### Efficient Syncing and Backups

Have you ever reorganized a photo library and then had to wait for rsync to re-upload gigabytes of data, even though the files themselves didn't change? With TooBig, you only sync the tiny *ref* files. The large *blobs* (your actual data) wasn't renamed or moved, making synchronization and backups dramatically faster and more efficient.

### Verifiable Data Integrity

Because each blob's filename is its SHA-256 checksum, verifying your entire collection is trivial. `toobig fsck` can easily confirm that none of your files have suffered from bit rot or corruption during transfers between hard drives, cloud storage, or backup media over the years.

### Version Control Your Files with Git

Git is powerful, but struggles with large binary files. TooBig allows you to commit your directory structure—the lightweight *ref* files—to a Git repository. This lets you track every change, rename, and reorganization of your assets without bloating the repository. You get a complete version history for your large files without the performance penalty of storing the actual files in Git.

### Data Deduplication

TODO: Need words here...
If you have the same file multiple times this will
store only one blob to represent all of those original files.


## Install

go install github.com/shortmoose/toobig[@version]


## Docs

### toobig update <repo.yaml>

Converts a set of files into a matching set of metadata files and a hardlink to its SHA256.

### toobig restore <repo.yaml>

Converts a set of metadata and SHA256 files into the original set of files.

### toobig fsck <repo.yaml>

Verifies the data integrity of a set of files, metadata files, and SHA256s to verify everything is consistent.

The -d \<dir\> doesn't use a repo.yaml file but instead just treats the directory as a blob directory to be verified.
