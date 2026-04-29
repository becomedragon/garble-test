// Copyright (c) 2020, The Garble Authors.
// See LICENSE for licensing information.

// A more complex main package that exercises symbol renaming, string constants,
// sub-package paths, and control-flow patterns — all key targets for Garble.

package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/becomedragon/garble-test/config"
	"github.com/becomedragon/garble-test/crypto"
	"github.com/becomedragon/garble-test/models"
	"github.com/becomedragon/garble-test/worker"
)

// globalVar is a package-level string literal — a target for string-constant obfuscation.
var globalVar = "global value"

// AuthService is a simple service type used to demonstrate reflection.
type AuthService struct{}

// LoginRequest holds credentials passed to authentication methods.
type LoginRequest struct {
	Email    string
	Password string
}

// ValidateToken returns true when the token equals the expected sentinel value.
func (AuthService) ValidateToken(token string) bool {
	return token == "ok"
}

// causePanic intentionally dereferences a nil *AuthService to trigger a panic,
// making stack-trace symbol readability observable before and after garble.
func causePanic() {
	var s *AuthService
	_ = s.ValidateToken("x")
}

// appVersion is another string constant in a different variable.
const appVersion = "1.0.0-garble"

// globalFunc is an exported-name target for symbol renaming.
func globalFunc() { fmt.Println("global func body") }

// generateID returns a random integer identifier.
func generateID() int {
	return rand.Intn(1_000_000)
}

// buildUsers creates a slice of demo users with nested control flow.
func buildUsers(count int) []*models.User {
	users := make([]*models.User, 0, count)
	for i := 0; i < count; i++ {
		id := generateID()
		name := fmt.Sprintf("User%d", id)
		email := fmt.Sprintf("user%d@example.com", i)
		u := models.NewUser(id, name, email)

		// Nested if / switch — CFG complexity.
		switch {
		case i%3 == 0:
			u.AddRole("admin")
		case i%3 == 1:
			u.AddRole("editor")
		default:
			u.AddRole("viewer")
		}

		if i%2 == 0 {
			u.AddRole("auditor")
		}

		users = append(users, u)
	}
	return users
}

// encodeUsers encodes each user's name with Base64 and the salted codec.
func encodeUsers(users []*models.User) []string {
	codec := crypto.NewSaltedCodec(crypto.NewBase64Codec())
	results := make([]string, len(users))
	for i, u := range users {
		results[i] = codec.Encode(u.Name)
	}
	return results
}

// verifyEncoding round-trips each encoded string through decode and checks it.
func verifyEncoding(encoded []string) {
	codec := crypto.NewSaltedCodec(crypto.NewBase64Codec())
	for i, e := range encoded {
		decoded, err := codec.Decode(e)
		if err != nil {
			fmt.Printf("  [!] decode error at index %d: %v\n", i, err)
			continue
		}
	// Nested if-else — CFG target. Guard against short decoded strings.
		if len(decoded) >= 4 {
			if decoded[:4] == "User" {
				fmt.Printf("  [ok] index %d: %s\n", i, decoded)
			} else {
				fmt.Printf("  [?]  index %d: unexpected prefix in %q\n", i, decoded)
			}
		} else {
			fmt.Printf("  [?]  index %d: decoded string too short: %q\n", i, decoded)
		}
	}
}

// runGoroutineFan demonstrates a fan-out / fan-in pattern with channels and select.
func runGoroutineFan(cfg *config.Config) {
	retries := cfg.GetInt(config.KeyRetries, 3)
	timeout := cfg.GetString(config.KeyTimeout, "5s")
	fmt.Printf("  fan-out with retries=%d timeout=%s\n", retries, timeout)

	type msg struct {
		id  int
		val string
	}

	ch := make(chan msg, retries)
	done := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < retries; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			encoded := crypto.EncodeData(fmt.Sprintf("payload-%d", idx))
			select {
			case ch <- msg{id: idx, val: encoded}:
			case <-done:
			}
		}(i)
	}

	// Close done after all goroutines finish.
	go func() {
		wg.Wait()
		close(done)
		close(ch)
	}()

	// Fan-in: collect results.
	for m := range ch {
		decoded, err := crypto.DecodeData(m.val)
		if err == nil {
			fmt.Printf("    goroutine %d → %s\n", m.id, decoded)
		}
	}
}

// runTaskPipeline builds tasks and processes them through the worker pool.
func runTaskPipeline() {
	titles := []string{
		"initialise database",
		"load configuration",
		"start HTTP server",
		"authenticate users",
		"process payments",
	}

	tasks := make([]*models.Task, len(titles))
	for i, title := range titles {
		tasks[i] = models.NewTask(i+1, title, i%3+1)
	}

	report := worker.RunTaskPipeline(tasks, 3)
	fmt.Println(" ", report.Summary())
}

// obfuscationDemo shows ObfuscateString / DeobfuscateString round-trip.
func obfuscationDemo() {
	secrets := []string{
		"top-secret-key",
		appVersion,
		globalVar,
	}
	for _, s := range secrets {
		obf := crypto.ObfuscateString(s)
		plain := crypto.DeobfuscateString(obf)
		fmt.Printf("  %q → obfuscated → %q\n", plain, obf)
	}
}

// reflectDemo exercises reflect.TypeOf, reflect.ValueOf, FieldByName, and
// MethodByName using models.User and models.Task.
// Before garble: output shows real type/method/field names.
// After garble: names become short random identifiers, so MethodByName
// lookups by the original name return an invalid Value.
func reflectDemo() {
	u := models.NewUser(42, "Alice", "alice@example.com")
	u.AddRole("admin")

	rv := reflect.ValueOf(u)
	rt := reflect.TypeOf(u)

	fmt.Printf("  reflect.TypeOf : %v\n", rt)
	fmt.Printf("  reflect.ValueOf: %v\n", rv)

	// FieldByName — after garble these exported field names are renamed and
	// FieldByName returns the zero Value (IsValid() == false).
	for _, field := range []string{"ID", "Name", "Email"} {
		f := rv.Elem().FieldByName(field)
		if f.IsValid() {
			fmt.Printf("  FieldByName(%q) = %v\n", field, f)
		} else {
			fmt.Printf("  FieldByName(%q): not found (garbled?)\n", field)
		}
	}

	// MethodByName — after garble "String" and "HasRole" are renamed, so
	// the Call below will be skipped and the fallback message printed.
	for _, call := range []struct {
		method string
		args   []reflect.Value
	}{
		{"String", nil},
		{"HasRole", []reflect.Value{reflect.ValueOf("admin")}},
	} {
		m := rv.MethodByName(call.method)
		if m.IsValid() {
			result := m.Call(call.args)
			fmt.Printf("  MethodByName(%q).Call() = %v\n", call.method, result[0])
		} else {
			fmt.Printf("  MethodByName(%q): not found (garbled?)\n", call.method)
		}
	}

	// Type name via reflect - garble replaces the package-qualified name.
	t := models.NewTask(1, "deploy", 2)
	rt2 := reflect.TypeOf(t).Elem()
	fmt.Printf("  Task type name : %s\n", rt2.Name())
	for _, fname := range []string{"ID", "Title", "Priority"} {
		sf, ok := rt2.FieldByName(fname)
		if ok {
			fmt.Printf("  Task field %q tag=%q\n", sf.Name, sf.Tag)
		} else {
			fmt.Printf("  Task field %q: not found (garbled?)\n", fname)
		}
	}
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	fmt.Println("=== garble-test", appVersion, "===")
	fmt.Println(globalVar)
	globalFunc()

	// ── Sub-package: config ──────────────────────────────────────────────────
	fmt.Println("\n[config]")
	cfg := config.New()
	cfg.Set(config.KeyDebug, true)
	fmt.Println(" ", cfg.Summary())

	override := config.New()
	override.Set(config.KeyRetries, 5)
	cfg.Merge(override)
	fmt.Println("  after merge:", cfg.Summary())

	// ── Sub-package: models ──────────────────────────────────────────────────
	fmt.Println("\n[models]")
	users := buildUsers(6)
	for _, u := range users {
		fmt.Println(" ", u)
	}

	// ── Sub-package: crypto ──────────────────────────────────────────────────
	fmt.Println("\n[crypto] encode/decode round-trip")
	encoded := encodeUsers(users)
	verifyEncoding(encoded)

	fmt.Println("\n[crypto] obfuscation demo")
	obfuscationDemo()

	// ── Goroutine fan-out / fan-in ───────────────────────────────────────────
	fmt.Println("\n[goroutines] fan-out/fan-in")
	runGoroutineFan(cfg)

	// ── Sub-package: worker (task pipeline) ──────────────────────────────────
	fmt.Println("\n[worker] task pipeline")
	runTaskPipeline()

	// ── Reflection demo ──────────────────────────────────────────────────────
	// Run before and after `garble build` to compare visible names.
	fmt.Println("\n[reflect] reflection calls demo")
	reflectDemo()

	// ── net/http (kept from original) ────────────────────────────────────────
	// These calls exercise the http symbol targets just as the original did.
	_ = http.Client{Transport: nil}

	fmt.Println("\nDone.")

	// ── AuthService reflection (explicit + robust) ────────────────────────────
	// Store the type once and guard method access so the code stays correct
	// even if the method set changes (e.g. after garble obfuscation).
	fmt.Println("\n[reflect] AuthService type/method names")
	t := reflect.TypeOf(AuthService{})
	fmt.Println("Type:", t.String())
	if t.NumMethod() > 0 {
		// Method(0) returns the first method in alphabetical order; printing
		// it confirms the name is readable before garble obfuscates it.
		fmt.Println("Method:", t.Method(0).Name)
	} else {
		fmt.Println("Method: <none>")
	}

	// ── Panic trigger (stack-trace readability) ───────────────────────────────
	causePanic()
}
