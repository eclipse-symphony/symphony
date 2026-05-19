#!/bin/bash
a=$1
echo $(echo $(<$1) | jq '.solutionversion.components') > ${a%.*}-output.${a##*.}