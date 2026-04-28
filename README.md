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
| **Reflection** — `TypeOf`, `ValueOf`, `FieldByName`, `MethodByName` | `main.go` (`reflectDemo`) |

## Comparing reflection output before vs after garble

The `reflectDemo` function in `main.go` is specifically designed to highlight what garble changes in the reflection output.

```sh
# 1. Build and run WITHOUT obfuscation — real names visible
go build -o garble-test-plain .
./garble-test-plain 2>&1 | grep -A 20 '\[reflect\]'

# 2. Build and run WITH garble obfuscation — names replaced
garble build -o garble-test-obfuscated .
./garble-test-obfuscated 2>&1 | grep -A 20 '\[reflect\]'
```

**Expected differences in the `[reflect]` section:**

| Output line | Plain binary | Garbled binary |
|---|---|---|
| `reflect.TypeOf` | `*models.User` | short random name |
| `FieldByName("Name")` | `Alice` | `not found (garbled?)` |
| `MethodByName("String").Call()` | user string | `not found (garbled?)` |
| `Task type name` | `Task` | random identifier |
| `Task field "Title"` | `Title` | `not found (garbled?)` |