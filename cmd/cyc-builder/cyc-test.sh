#!/bin/bash

# get absolute path
WD=`echo $(pwd)/$line`

ROOT=$WD/cyclus-full
SRC=$ROOT/src
INSTALL=$ROOT/install
URL=http://github.com/cyclus

# test the installation
echo "Running cyclus tests:"
results=$INSTALL/cyclus/bin/CyclusUnitTestDriver
echo $results

# indicated failed tests in stderr
if [[ `echo $results | grep FAIL` ]]; then
  echo "failed" 1>&2
fi

# cleanup
cd $ROOT/..
rm -Rf $ROOT

