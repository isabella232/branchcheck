Branchcheck
===========

A git-related tool in support of gitflow branch naming to verify
the branch name is consistent with the stated artifact version in
all POMs in a tree.  Versions on branch develop must be named
major.minor-SNAPSHOT.  Feature branches follow jgitflow naming:
major.minor-xx_dddd-SNAPSHOT.  This check is enabled by default.

Branchcheck will also iterate over a repository's branches and verify that no two
branches have the same stated Maven POM version.  This check is enabled with the -version-dups switch.

Go install
==========

Quick install:  http://goo.gl/Or96vH

The longer version:  http://golang.org/doc/install

Install Branchcheck
===================

To just install and run branchcheck

     $ go get github.com/xoom/branchcheck
     $ which branchcheck
     /Users/mpetrovic/Projects/go/bin/branchcheck

Build
=====

If you want to modify branchcheck

     $ cd $GOPATH/src/github.com/xoom/branchcheck
     $ go test
     $ go build # or make

Branch name / POM version verification
======================================

Verify that for the current branch, the branch name and Maven POM versions are consistent.

There may be POMs that you wish to exclude from processing.  Notate these with the -excludes command 
line switch

     branchcheck -excludes apath/pom.xml,bpath/pom.xml

Verify that no two branches claim the same Maven artifact version in the top level POM
======================================================================================

     branchcheck -version-dups

Branch duplicate verification requires network access to the Git remote for purposes of running git-fetch and git-ls-remote.

Print the version claimed by the pom.xml file in the current working directory
==============================================================================

     branchcheck -pom-version

Logging
=======

Set the debug flag to debug:

     $ branchcheck -debug
     2014/10/15 17:28:21 git [rev-parse --abbrev-ref HEAD]
     2014/10/15 17:28:21 Validating branch master
     2014/10/15 17:28:21 Analyzing pom.xml
     2014/10/15 17:28:21 master is a feature branch: false

Help
====

     $ branchcheck -h
