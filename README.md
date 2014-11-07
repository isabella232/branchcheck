Branchcheck
===========

A git pre-commit hook in support of jgitflow branch naming to verify
the branch name is consistent with the stated artifact version in
all POMs in a tree.  Versions on branch develop must be named
major.minor-SNAPSHOT.  Feature branches follow jgitflow naming:
major.minor-xx_dddd-SNAPSHOT.

Go install
==========

Quick install:  http://goo.gl/Or96vH

The longer version:  http://golang.org/doc/install

Install Branchcheck
===================

     $ go get github.com/xoom/branchcheck
     $ which branchcheck
     /Users/mpetrovic/Projects/go/bin/branchcheck

Build
=====

     $ cd $GOPATH/src/github.com/xoom/branchcheck
     $ go test
     $ go build # or make

Exclusions
==========

There may be POMs that you wish to exclude from processing.  Notate these with the -excludes command 
line switch

     branchcheck -excludes apath/pom.xml,bpath/pom.xml

Logging
=======

Set the BRANCHCHECK_DEBUG environment variable to "true" for some debug:

     $ BRANCHCHECK_DEBUG=true branchcheck
     2014/10/15 17:28:21 git [rev-parse --abbrev-ref HEAD]
     2014/10/15 17:28:21 Validating branch master
     2014/10/15 17:28:21 Analyzing pom.xml
     2014/10/15 17:28:21 master is a feature branch: false
