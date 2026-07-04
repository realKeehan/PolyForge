package kumi

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"unicode/utf16"
)

// ── Windows .lnk (shell link) parser ─────────────
// Minimal pure-Go reader for the MS-SHLLINK binary format so shortcuts can
// be resolved without COM. Extracts the link target path and command-line
// arguments, which is all the launcher search needs.

const lnkHeaderSize = 0x4C

// Link flags (MS-SHLLINK 2.1.1)
const (
	lnkHasLinkTargetIDList = 0x01
	lnkHasLinkInfo         = 0x02
	lnkHasName             = 0x04
	lnkHasRelativePath     = 0x08
	lnkHasWorkingDir       = 0x10
	lnkHasArguments        = 0x20
	lnkIsUnicode           = 0x80
)

var errNotShellLink = errors.New("not a shell link file")

// parseShortcut reads a .lnk file and returns its target path and arguments.
// The target comes from LinkInfo's LocalBasePath when present, falling back
// to the relative path resolved against the shortcut's own directory.
func parseShortcut(path string) (target string, args string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	if len(data) < lnkHeaderSize || binary.LittleEndian.Uint32(data[0:4]) != lnkHeaderSize {
		return "", "", errNotShellLink
	}

	flags := binary.LittleEndian.Uint32(data[20:24])
	pos := lnkHeaderSize

	// LinkTargetIDList: length-prefixed, skipped entirely.
	if flags&lnkHasLinkTargetIDList != 0 {
		if pos+2 > len(data) {
			return "", "", errNotShellLink
		}
		pos += 2 + int(binary.LittleEndian.Uint16(data[pos:]))
	}

	// LinkInfo: carries the absolute local base path for local targets.
	if flags&lnkHasLinkInfo != 0 {
		if pos+4 > len(data) {
			return "", "", errNotShellLink
		}
		liSize := int(binary.LittleEndian.Uint32(data[pos:]))
		if liSize >= 28 && pos+liSize <= len(data) {
			target = linkInfoLocalPath(data[pos : pos+liSize])
		}
		pos += liSize
	}

	// StringData section: optional strings in a fixed order.
	unicode := flags&lnkIsUnicode != 0
	readString := func() string {
		if pos+2 > len(data) {
			pos = len(data)
			return ""
		}
		count := int(binary.LittleEndian.Uint16(data[pos:]))
		pos += 2
		if unicode {
			end := pos + count*2
			if end > len(data) {
				pos = len(data)
				return ""
			}
			u := make([]uint16, count)
			for i := 0; i < count; i++ {
				u[i] = binary.LittleEndian.Uint16(data[pos+i*2:])
			}
			pos = end
			return string(utf16.Decode(u))
		}
		end := pos + count
		if end > len(data) {
			pos = len(data)
			return ""
		}
		s := string(data[pos:end])
		pos = end
		return s
	}

	var relative string
	if flags&lnkHasName != 0 {
		_ = readString()
	}
	if flags&lnkHasRelativePath != 0 {
		relative = readString()
	}
	if flags&lnkHasWorkingDir != 0 {
		_ = readString()
	}
	if flags&lnkHasArguments != 0 {
		args = readString()
	}

	if target == "" && relative != "" {
		target = filepath.Clean(filepath.Join(filepath.Dir(path), relative))
	}
	if target == "" {
		return "", "", errors.New("shortcut has no resolvable target")
	}
	return target, args, nil
}

// linkInfoLocalPath extracts LocalBasePath + CommonPathSuffix from a LinkInfo
// block, preferring the Unicode variants when the header advertises them.
func linkInfoLocalPath(li []byte) string {
	headerSize := int(binary.LittleEndian.Uint32(li[4:8]))
	liFlags := binary.LittleEndian.Uint32(li[8:12])
	const volumeIDAndLocalBasePath = 0x1
	if liFlags&volumeIDAndLocalBasePath == 0 {
		return "" // network-relative link; not a local launcher install
	}

	base, suffix := "", ""
	if headerSize >= 0x24 && len(li) >= 36 {
		base = readUTF16At(li, int(binary.LittleEndian.Uint32(li[28:32])))
		suffix = readUTF16At(li, int(binary.LittleEndian.Uint32(li[32:36])))
	}
	if base == "" {
		base = readANSIAt(li, int(binary.LittleEndian.Uint32(li[16:20])))
		suffix = readANSIAt(li, int(binary.LittleEndian.Uint32(li[24:28])))
	}
	return base + suffix
}

func readANSIAt(buf []byte, off int) string {
	if off <= 0 || off >= len(buf) {
		return ""
	}
	end := off
	for end < len(buf) && buf[end] != 0 {
		end++
	}
	return string(buf[off:end])
}

func readUTF16At(buf []byte, off int) string {
	if off <= 0 || off+1 >= len(buf) {
		return ""
	}
	var u []uint16
	for i := off; i+1 < len(buf); i += 2 {
		c := binary.LittleEndian.Uint16(buf[i:])
		if c == 0 {
			break
		}
		u = append(u, c)
	}
	return string(utf16.Decode(u))
}
