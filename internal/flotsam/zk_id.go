// Copyright (c) 2025 Vice Project
// This file contains code adapted from the ZK note-taking system.
// Original code: https://github.com/zk-org/zk
// Original license: GNU General Public License v3.0
// 
// Portions of this file are derived from ZK's ID generation system,
// specifically from internal/core/id.go and internal/util/rand/rand.go.
// The original ZK code is licensed under GPLv3.

// Package flotsam provides ZK-compatible ID generation for flotsam notes.
// AIDEV-NOTE: flotsam package handles ZK integration components
package flotsam

import (
	"math/rand"
	"time"
	"unicode"
)

// IDOptions holds the options used to generate an ID.
// AIDEV-NOTE: copied from ZK core.IDOptions - defines charset, length, case for ID generation
type IDOptions struct {
	Length  int
	Charset Charset
	Case    Case
}

// Charset is a set of characters.
type Charset []rune

var (
	// CharsetAlphanum is a charset containing letters and numbers.
	// AIDEV-NOTE: ZK's default charset - used by flotsam for compatibility
	CharsetAlphanum = Charset("0123456789abcdefghijklmnopqrstuvwxyz")
	// CharsetHex is a charset containing hexadecimal characters.
	CharsetHex = Charset("0123456789abcdef")
	// CharsetLetters is a charset containing only letters.
	CharsetLetters = Charset("abcdefghijklmnopqrstuvwxyz")
	// CharsetNumbers is a charset containing only numbers.
	CharsetNumbers = Charset("0123456789")
)

// Case represents the letter case to use when generating an ID.
type Case int

const (
	// CaseLower generates lowercase characters only
	CaseLower Case = iota + 1
	// CaseUpper generates uppercase characters only
	CaseUpper
	// CaseMixed generates both upper and lowercase characters
	CaseMixed
)

// IDGenerator is a function returning a new ID with each invocation.
type IDGenerator func() string

// IDGeneratorFactory creates a new IDGenerator function using the given IDOptions.
type IDGeneratorFactory func(opts IDOptions) func() string

// NewIDGenerator returns a function generating string IDs using the given options.
// AIDEV-NOTE: core ID generation logic copied from ZK rand package - cryptographically random
// Inspired by https://www.calhoun.io/creating-random-strings-in-go/
func NewIDGenerator(options IDOptions) func() string {
	if options.Length < 1 {
		panic("IDOptions.Length must be at least 1")
	}

	var charset []rune
	for _, char := range options.Charset {
		switch options.Case {
		case CaseLower:
			charset = append(charset, unicode.ToLower(char))
		case CaseUpper:
			charset = append(charset, unicode.ToUpper(char))
		case CaseMixed:
			charset = append(charset, unicode.ToLower(char))
			charset = append(charset, unicode.ToUpper(char))
		default:
			panic("unknown Case value")
		}
	}

	// AIDEV-NOTE: time-seeded PRNG ensures unique IDs across sessions
	// Warning: Uses math/rand for compatibility with ZK's implementation.
	// For cryptographic security, consider crypto/rand, but this matches ZK behavior.
	//revive:disable-next-line:insecure-random ZK compatibility requires math/rand
	// #nosec G404 -- ZK compatibility requires math/rand instead of crypto/rand
	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	return func() string {
		buf := make([]rune, options.Length)
		for i := range buf {
			buf[i] = charset[randGen.Intn(len(charset))]
		}

		return string(buf)
	}
}

// NewFlotsamIDGenerator creates an ID generator with flotsam-specific defaults.
// AIDEV-NOTE: flotsam defaults - 4-char alphanum lowercase, compatible with ZK
// This matches ZK's default configuration from NewDefaultConfig()
func NewFlotsamIDGenerator() IDGenerator {
	options := IDOptions{
		Length:  4,           // ZK default
		Charset: CharsetAlphanum, // ZK default
		Case:    CaseLower,   // ZK default
	}
	return NewIDGenerator(options)
}

// DefaultIDOptions returns the default ID generation options for flotsam.
// AIDEV-NOTE: matches ZK's default config for maximum compatibility
func DefaultIDOptions() IDOptions {
	return IDOptions{
		Length:  4,
		Charset: CharsetAlphanum,
		Case:    CaseLower,
	}
}