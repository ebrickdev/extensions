#!/bin/bash

# Find all directories containing a go.mod file and run go mod tidy
find . -name "go.mod" -execdir go mod tidy \;

echo "go mod tidy has been run in all directories containing a go.mod file."