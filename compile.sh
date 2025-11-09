#!/bin/bash

# The -tags "release" flag tells the compiler to use the "path_prod.go" file
go build -o ./bin/flick -tags "release" ./main.go
