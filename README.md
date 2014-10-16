Branchcheck
===========

A git pre-commit hook to verify the branch name is consistent with
the stated artifact version in the POMs.  A git flow thing.

Build
=====

     go build
     go test

Then copy ./branchcheck into .git/hooks/pre-commit for local repos of interest.

You can also just run branchcheck at the top level of your repository,
which is exactly where it would run if it were a pre-commit hook.

Logging
=======

Set the BRANCHCHECK_DEBUG environment variable to "true" for some debug:

     $ BRANCHCHECK_DEBUG=true branchcheck
     2014/10/15 17:28:21 git [rev-parse --abbrev-ref HEAD]
     2014/10/15 17:28:21 Validating branch master
     2014/10/15 17:28:21 Analyzing pom.xml
     2014/10/15 17:28:21 master is a feature branch: false
