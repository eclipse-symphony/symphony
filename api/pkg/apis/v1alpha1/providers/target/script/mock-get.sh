#!/bin/bash
a=$1
echo $(echo $(<$1) | jq '.stages[0].solution.components') > ${a%.*}-output.${a##*.}