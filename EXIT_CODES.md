# Exit Codes

0 - Success

1 - General Error

2 - Panic, application crashed - golang defined exit code

3 - Invalid arguments, invalid command, etc.

10+ - These are what we will be using for specific exit codes.

10 - normal operational updates - for example a file has been added, etc. These
normally exhibit as success (exit 0) but exit code 10 is used with the flag
--ten.

11 - data inconsistencies (corrupted blob, invalid ref, etc) - status, update,
restore, fsck.

12 - problem found with the toobig config file.

100+ - We aren't using anything up here.
