# TooBig

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/shortmoose/toobig)
[![Go Report Card](https://goreportcard.com/badge/shortmoose/toobig)](https://goreportcard.com/report/shortmoose/toobig)
[![Releases](https://img.shields.io/github/release-pre/shortmoose/toobig.svg?sort=semver)](https://github.com/shortmoose/toobig/releases)
[![LICENSE](https://img.shields.io/github/license/shortmoose/toobig.svg)](https://github.com/shortmoose/toobig/blob/master/LICENSE)

TooBig was created from a need to manage largish files like photos, videos, and
binaries. Over the last number of years I have found it has worked wonderfully
for managing my decades of family photos, but also for managing photo and video
assets for development projects, and other uses here and there.

The basics of what TooBig does is it converts a set of files
into a set of refs and blobs. A blob is the original
file, but the name has been changed to the 64 character SHA-256 checksum of the
contents of the file. A ref is a small file that has the same name and directory
structure of the original file, but the file itself is tiny, containing just the
SHA-256 checksum of the original file and a timestamp.

`toobig update` converts files to refs/blobs.
`toobig restore` converts refs/blobs to files.

## Why TooBig

- rsync'ing files between two computers. Normally if you change directory
  structure and names of files rsyncing becomes much more expensive. With toobig
  the filename structure is represented by small files (cheap to rsync), the
  blobs don't move or get renamed.
- Backing up files is simpler and cheaper for the same reasons that rsync
  is nicer.
- Easy to validate file integrity. Since every blob has its checksum as the
  filename it is straightforward to verify none of the files have been
  corrupted over years of moving them between different computers, hard drives,
  backups, etc.
- Easy to use a version management tool like `git` to store the refs in. This
  allows you to track all the changes to your files without actually storing
  your actual data in `git`. (Git really doesn't do well with GBs of large
  files, that is what I used to do...)
- Data deduplication. If you have the same file multiple times this will
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
