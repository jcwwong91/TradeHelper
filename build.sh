#!/bin/bash

# Useful if developing on windows

set -e
export GOOS=linux
export GOARCH=amd64
go build
docker build -t tradehelper .
