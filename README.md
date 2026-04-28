# garble-test

A simple Go project to test [Garble](https://github.com/burrowers/garble)'s obfuscation capability.

## Prerequisites

- [Go](https://golang.org/dl/) 1.18+
- [Garble](https://github.com/burrowers/garble): `go install mvdan.cc/garble@latest`

## Build normally

```sh
go build .
```

## Build with Garble obfuscation

```sh
garble build .
```

## Run

```sh
./garble-test
```

## What's included

- Global variable (`globalVar`)
- Global function (`globalFunc`)
- Usage of `net/http` package (both server and client side)
- A mix of funcs, methods, fields, and string literals — good targets for Garble obfuscation