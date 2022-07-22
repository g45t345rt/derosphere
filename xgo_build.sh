#!/bin/bash

echo -n "Enter START_ENV: " # simulator, testnet, mainnet
read env
flags="-X github.com/g45t345rt/derosphere/config.START_ENV=$env"
xgo -v -ldflags "$flags" -targets windows/amd64,windows/386,linux/amd64,linux/386,linux/arm64,darwin/amd64,darwin/arm64 .