#! /usr/bin/env bash

set -e

if ! [[ -x "$(command -v protoc-gen-go)" ]]; then
  echo "Need to install protoc-gen-go"
  exit 1
fi

PROTOC_OPTS="-I./apis/ --go_out=temp --go-grpc_out=temp"

mkdir -p temp/

protoc $PROTOC_OPTS ./apis/teams_voicein/*.proto

rm -rf pkg/teams_voicein

mv temp/github.com/kanataidarov/tinkoff_voicekit/pkg/* pkg

rm -rf temp/
