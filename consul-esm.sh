#! /usr/bin/env bash
# consul-esm.sh

cd ./consul-esm

go mod tidy


# make -C build
# make -C test

# go env


export CONSUL_ESM_VERSION="0.3.1"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X github.com/hashicorp/consul-esm/version.Name="consul-esm" -X github.com/hashicorp/consul-esm/version.Version="$CONSUL_ESM_VERSION"" -o "dist/linux/amd64/consul-esm-linux-$CONSUL_ESM_VERSION"

ls -l dist/linux/amd64/consul-esm*


go test ./
go test -v ./...

