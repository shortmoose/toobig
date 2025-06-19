# TooBig

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/shortmoose/toobig)
[![Go Report Card](https://goreportcard.com/badge/shortmoose/toobig)](https://goreportcard.com/report/shortmoose/toobig)
[![Releases](https://img.shields.io/github/release-pre/shortmoose/toobig.svg?sort=semver)](https://github.com/shortmoose/toobig/releases)
[![LICENSE](https://img.shields.io/github/license/shortmoose/toobig.svg)](https://github.com/shortmoose/toobig/blob/master/LICENSE)

TooBig was created from a need to manage large files like photos, videos, and
other files. Over the last number of years I have found it has worked
wonderfully for managing my decades of family photos, but also for managing
photo and video assets for development projects, and other uses here and there.

TooBig converts a set of files into a set of refs and blobs. A blob is the
original file, but the name has been changed to the 64 character SHA-256
checksum of the contents of the file. A ref is a small file that has the same
name and directory structure of the original file, but the file itself is tiny,
containing just the SHA-256 checksum of the original file and a timestamp.

- `toobig update` converts files to refs/blobs.
- `toobig restore` converts refs/blobs to files.

## Why Use TooBig?

TooBig's architecture of separating file metadata (refs) from file content
(blobs) provides several advantages:

### Efficient Syncing and Backups

Have you ever reorganized a photo library and then had to wait for rsync to
re-upload gigabytes of data, even though the files themselves didn't change?
With TooBig, only the tiny *ref* files would need to be re-uploaded. The large
*blobs* (your actual data) wasn't renamed or moved, making synchronization and
backups dramatically faster and more efficient.

### Verifiable Data Integrity

Because each blob's filename is its SHA-256 checksum, verifying your entire
collection is trivial. `toobig fsck` can easily confirm that none of your files
have suffered from bit rot or corruption during transfers between hard drives,
cloud storage, or backup media over the years.

### Version Control Your Files with Git

Git is powerful, but struggles with large binary files. TooBig allows you to
commit your directory structure—the lightweight *ref* files—to a Git
repository. This lets you track every change, rename, and reorganization of
your files without bloating the repository. You get a complete version history
for your large files without the performance penalty of storing the actual
data in Git.

### Data Deduplication

*TODO:* Need words here... If you have the same file multiple times this will
store only one blob to represent all of those original files.


## Install

go install github.com/shortmoose/toobig@latest


## Usage

### toobig update <repo.cfg>

Converts a set of files into a matching set of refs and blobs.

### toobig restore <repo.cfg>

Converts a set of refs and blobs into a matching set of files.

### toobig status <repo.cfg>

Which files need to be updated. (Basically a dry-run for `toobig update`.)

### toobig fsck <repo.cfg>

Verifies the integrity of the refs and blobs.


## Exit Codes

Some effort has been made to make error codes consistent and useful. Here is a
basic list of the error codes we currently use.

- **0** - Success.
- **1** - General error.
- **2** - Panic, application crashed - golang defined exit code.
- **3** - Invalid arguments, invalid command, etc.
- **10** - Normal operational updates - for example a file has been added, etc.
  These normally exhibit as success (exit 0) but exit code 10 is used with the
  flag --update-is-error.
- **11** - Data inconsistencies, generally will need manual intervention to
  repair (corrupted blob, invalid ref, etc).
- **12** - Error with the config file.
- **125+** - Above this range is usually used for signal handling. For example
  a ctrl-c will exit with the code 130.
