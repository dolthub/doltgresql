#!/bin/bash

# Set the working directory to the directory of the script's location
cd "$(cd -P -- "$(dirname -- "$0")" && pwd -P)"
cd ..

go run honnef.co/go/tools/cmd/staticcheck@2025.1 ./...
