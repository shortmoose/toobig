# Exit Codes

0 - Success

1 - General Error

2 - Panic, application crashed - golang defined exit code

3 - Invalid arguments, invalid command, etc.

10+ - These are what we will be using for specific exit codes.

10 - status found changes needing update, fsck found problems.

11 - problem found with the toobig config file.

12 - status and update - files need or were updated.

13 - status and update - errors that couldn't be updated.

100+ - We aren't using anything up here.
