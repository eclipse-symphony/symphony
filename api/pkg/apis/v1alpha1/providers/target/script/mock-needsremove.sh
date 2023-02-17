#!/bin/bash
current=$(echo $(<$1) | jq '.')
desired=$(echo $(<$2) | jq '.')
echo "false"