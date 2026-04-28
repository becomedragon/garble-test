// Package crypto provides string encoding, decoding, and simple obfuscation helpers.
package crypto

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

const (
	defaultSalt   = "g4rbl3-s4lt"
	encodingLabel = "base64"
)

// Encoder is the interface for types that can encode a string.
type Encoder interface {
	Encode(data string) string
}

// Decoder is the interface for types that can decode a string.
type Decoder interface {
	Decode(data string) (string, error)
}

// Codec combines Encoder and Decoder.
type Codec interface {
	Encoder
	Decoder
}

// Base64Codec implements Codec using standard Base64 encoding.
type Base64Codec struct {
	label string
}

// NewBase64Codec returns a Base64Codec ready for use.
func NewBase64Codec() *Base64Codec {
	return &Base64Codec{label: encodingLabel}
}

// Encode encodes data to a Base64 string.
func (c *Base64Codec) Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// Decode decodes a Base64 string back to its original form.
func (c *Base64Codec) Decode(data string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	return string(b), nil
}

// Label returns the encoding scheme label.
func (c *Base64Codec) Label() string { return c.label }

// SaltedCodec wraps a Codec and prepends/removes a salt marker.
type SaltedCodec struct {
	inner Codec
	salt  string
}

// NewSaltedCodec creates a SaltedCodec around the given Codec.
func NewSaltedCodec(inner Codec) *SaltedCodec {
	return &SaltedCodec{inner: inner, salt: defaultSalt}
}

// Encode encodes data and prepends the salt marker.
func (s *SaltedCodec) Encode(data string) string {
	encoded := s.inner.Encode(data)
	return s.salt + ":" + encoded
}

// Decode removes the salt marker and decodes the data.
func (s *SaltedCodec) Decode(data string) (string, error) {
	prefix := s.salt + ":"
	if !strings.HasPrefix(data, prefix) {
		return "", errors.New("crypto: missing salt prefix")
	}
	return s.inner.Decode(strings.TrimPrefix(data, prefix))
}

// EncodeData is a convenience function that encodes data using Base64.
func EncodeData(data string) string {
	c := NewBase64Codec()
	return c.Encode(data)
}

// DecodeData is a convenience function that decodes a Base64 string.
func DecodeData(data string) (string, error) {
	c := NewBase64Codec()
	return c.Decode(data)
}

// ObfuscateString applies a simple XOR-like substitution to make a string less
// obvious in binaries. This is intentionally weak and for demonstration only.
func ObfuscateString(s string) string {
	runes := []rune(s)
	for i, r := range runes {
		runes[i] = r ^ 0x5A
	}
	return string(runes)
}

// DeobfuscateString reverses ObfuscateString.
func DeobfuscateString(s string) string {
	return ObfuscateString(s) // XOR is its own inverse
}
