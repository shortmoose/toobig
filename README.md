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


## Warnings

Please realize that TooBig leverages hard links for much of its functionality.
Make sure you realize what this means. Specifically you shouldn't edit files in
place, this will make the linked blob file no longer match its checksum
(filename). Files should break the hard link when they are edited. If you are
unsure what the software you use does, make sure you run `toobig fsck`
regularly to help you validate everything is working correctly.


## Install

go install github.com/shortmoose/toobig@latest


## Usage

Basic Usage:

```bash
# Dump photos from SD card to my photo directory.
# Delete/Edit photos as needed.
toobig status photo-toobig.cfg  # Not actually necessary
toobig update photo-toobig.cfg
rsync [flags] {refs,blobs} cloud-backup-dir
```

An example set up. Sort of git-esque:

```bash
cd <photos-dir>
mkdir .toobig
mkdir .toobig/{refs,blobs,dups}
toobig config >.toobig/config
# Edit config to look like:
# {
#  "file-path": "..",
#  "blob-path": "blobs",
#  "ref-path": "refs",
#  "dup-path": "dups"
# }
#
# You can verify your config by running:
toobig status .toobig/config
```

*TODO:* Give more examples here...


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


## Future

- Currently we never delete old blobs, which means I normally run a bash script
  that moves all blobs with only 1 link, ie `find -links 1 -type f`, to a
  different directory, then run `toobig fsck`, to verify I didn't remove
  anything I shouldn't have. This functionality should be part of the `update`
  command.
- The output, especially with errors, could be more consistent.
- I have thought about breaking up the blobs directory into subdirectories,
  sort of like git does with its objects directory. No immediate plans for this
  I would need to see people using it with enough blobs to make it worth while.
- A config option to *NOT* use hardlinks? It would double the space usage but
  allow people to us work flows where they actually modify files, instead of
  replace them.
- Other ideas??
