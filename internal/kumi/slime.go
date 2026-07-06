package kumi

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
)

// ══════════════════════════════════════════════════
// .slime container format
//
// A .slime file is PolyForge's branded pack container: a standard ZIP
// archive whose bytes are obfuscated so the file is not a plain openable
// zip and gets its own extension + double-click handler.
//
// IMPORTANT: this is obfuscation, not cryptography. The keystream is
// derived from a constant, so anyone with this code can reverse it. Its
// job is format obscurity and branding, not access control — real access
// control for private packs is the server-side password gate
// (api/pack-access.php + VerifyPackAccess).
//
// Layout:
//   [0:5]  magic  "SLIME"
//   [5]    version 0x01
//   [6]    flags   0x00 (0x00 = xor-obfuscated zip payload)
//   [7]    reserved 0x00
//   [8:]   zip bytes, transformed by slimeTransform
//
// The transform is symmetric and byte-wise, so the exact same routine is
// reproduced in the PowerShell packager and the PHP admin packager:
//   out[i] = in[i] XOR key[i % 32] XOR (i & 0xFF)
//   key    = SHA-256("PolyForge-Slime-v1")   (32 bytes)
// ══════════════════════════════════════════════════

// ── Future reference: LZMA payload compression (heavy update) ──────
// The .slime payload is a DEFLATE zip today. LZMA (xz) compresses better on
// low-entropy data; for modpacks the real wins are Distant Horizons LOD
// databases and uncompressed resource/datapacks. Already-compressed .jar /
// .zip / .png files gain little (that's most of a typical pack's bytes),
// so pair it with store-mode for those and LZMA only the rest.
//
// Measured cost to add a pure-Go, CGO-free codec (keeps our clean
// GOOS=linux/darwin cross-compiles), stripped binary delta over a baseline:
//   github.com/ulikunitz/xz    raw xz/LZMA codec    +~284 KB, 1 dep    ← preferred
//   github.com/bodgit/sevenzip full 7z container    +~3.5 MB, ~13 deps ← avoid
// We own both ends of .slime, so the 7z *container* is unnecessary — a
// future update would xz-compress the payload with ulikunitz/xz, set a new
// slimeFlags value to signal the codec, and keep flags 0x00 (DEFLATE)
// readable for back-compat. Bandwidth-wise, delta updates (ComparePackMods
// on the manifest hashes) still save far more than any codec swap.

const (
	slimeVersion  = 0x01
	slimeFlagsXor = 0x00 // 0x00 = XOR-obfuscated DEFLATE zip; reserve others for LZMA
	slimeKeyPhrase = "PolyForge-Slime-v1"
)

var slimeMagic = []byte("SLIME")

// slimeKey returns the 32-byte obfuscation key. Kept as a function so the
// derivation stays identical to the other language implementations.
func slimeKey() [32]byte {
	return sha256.Sum256([]byte(slimeKeyPhrase))
}

// slimeTransform applies the symmetric XOR obfuscation in place-safe form,
// returning a new slice. Running it twice returns the original bytes.
func slimeTransform(payload []byte) []byte {
	key := slimeKey()
	out := make([]byte, len(payload))
	for i, b := range payload {
		out[i] = b ^ key[i%32] ^ byte(i)
	}
	return out
}

// WrapSlime turns raw zip bytes into a .slime container.
func WrapSlime(zipBytes []byte) []byte {
	header := []byte{slimeMagic[0], slimeMagic[1], slimeMagic[2], slimeMagic[3], slimeMagic[4], slimeVersion, slimeFlagsXor, 0x00}
	return append(header, slimeTransform(zipBytes)...)
}

// IsSlime reports whether data starts with the .slime magic header.
func IsSlime(data []byte) bool {
	return len(data) >= 8 && bytes.Equal(data[:5], slimeMagic)
}

// UnwrapSlime reverses WrapSlime, returning the underlying zip bytes.
func UnwrapSlime(data []byte) ([]byte, error) {
	if !IsSlime(data) {
		return nil, fmt.Errorf("not a .slime file (bad magic)")
	}
	if data[5] != slimeVersion {
		return nil, fmt.Errorf("unsupported .slime version %d", data[5])
	}
	if data[6] != slimeFlagsXor {
		return nil, fmt.Errorf("unsupported .slime flags 0x%02x", data[6])
	}
	return slimeTransform(data[8:]), nil
}

// readPackArchive reads a pack file from disk and returns the plaintext zip
// bytes. It accepts both .slime containers and plain .zip/.polypack.zip
// files (detected by magic), so dev builds and end users interoperate.
func readPackArchive(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if IsSlime(data) {
		return UnwrapSlime(data)
	}
	return data, nil // already a plain archive
}
