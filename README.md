# garble-test

A Go project to test [Garble](https://github.com/burrowers/garble)'s obfuscation capability across a wide range of language features.

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

## Project layout

```
garble-test/
├── main.go          # Entry point — complex control flow, goroutines, fan-out/fan-in
├── models/          # Structs, interfaces, exported & unexported symbols
│   └── models.go
├── crypto/          # Base64 codec, salted codec, string obfuscation helpers
│   └── crypto.go
├── config/          # Thread-safe config with typed accessors and merge
│   └── config.go
└── worker/          # Goroutine pool, job/result channels, select statements
    └── worker.go
```

## What's covered for Garble obfuscation

| Garble target | Where |
|---|---|
| **Symbol renaming** — exported & unexported funcs, types, fields | `models`, `config`, `crypto`, `worker`, `main` |
| **String constants** — literals, format strings, Base64 encode/decode, XOR obfuscation | `crypto/crypto.go`, `main.go` |
| **Package paths** — multiple sub-packages under one module | `models/`, `config/`, `crypto/`, `worker/` |
| **CFG complexity** — deep if/else, switch, for, goroutines, channels, select | `main.go`, `worker/worker.go` |