#!/bin/bash
a=$1
echo $(echo $(<$1) | jq '.solution.components') > ${a%.*}-output.${a##*.}