Branchcheck
===========

A git pre-commit hook to verify the branch name is consistent with
the stated artifact version in the POMs.  A git flow thing.

Build
=====

go build
go test

Then copy ./branchcheck into .git/hooks/pre-commit for local repos of interest.
