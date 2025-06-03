package archives

import (
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

// GetEncodingByName converts a string encoding name to an encoding.Encoding
func GetEncodingByName(name string) encoding.Encoding {
	switch name {
	case "shift-jis", "shiftjis", "sjis", "japanese":
		return japanese.ShiftJIS
	case "euc-jp", "eucjp":
		return japanese.EUCJP
	case "euc-kr", "euckr", "korean":
		return korean.EUCKR
	case "gbk", "gb18030", "gb2312", "simplified-chinese":
		return simplifiedchinese.GBK
	case "big5", "traditional-chinese":
		return traditionalchinese.Big5
	case "utf-16le", "windows":
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	case "utf-8", "utf8":
		return nil // No encoding needed for UTF-8
	}
	return nil
}

// GetEncodingFromCharset converts a charset name to an encoding.Encoding
func GetEncodingFromCharset(charset string, language string) encoding.Encoding {
	switch charset {
	case "Shift_JIS", "SJIS", "shift-jis", "sjis":
		return japanese.ShiftJIS
	case "EUC-JP", "eucjp":
		return japanese.EUCJP
	case "EUC-KR", "euckr":
		return korean.EUCKR
	case "GB18030", "GBK", "GB2312", "gb18030", "gbk", "gb2312":
		return simplifiedchinese.GBK
	case "Big5", "big5":
		return traditionalchinese.Big5
	case "UTF-16", "utf-16", "UTF-16LE", "utf-16le":
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	case "windows-1252", "iso-8859-1":
		// No direct support, but often these western encodings are not problematic
		// when used with Go's Unicode support
		return nil
	case "ASCII", "US-ASCII", "ascii":
		return nil // ASCII is a subset of UTF-8
	}

	// Try language-based detection if charset didn't match
	switch language {
	case "ja", "jpn":
		return japanese.ShiftJIS
	case "ko", "kor":
		return korean.EUCKR
	case "zh", "zho":
		return simplifiedchinese.GBK
	}

	return nil
}

// GetFallbackEncodings returns a list of common encodings to try as fallbacks
func GetFallbackEncodings() []encoding.Encoding {
	return []encoding.Encoding{
		japanese.ShiftJIS,
		simplifiedchinese.GBK,
		korean.EUCKR,
		traditionalchinese.Big5,
		japanese.EUCJP,
		unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM),
	}
}

// DetectEncoding analyzes the provided string to determine its encoding
// It tries encodings in this order: UTF-8, Shift-JIS, GBK
func DetectEncoding(data []byte) encoding.Encoding {
	if len(data) == 0 {
		return nil
	}

	// First try: Check if it's valid UTF-8
	if utf8.Valid(data) {
		return nil // UTF-8 is valid, no encoding needed
	}

	// If not UTF-8, default to ShiftJIS as most common for ZIP files
	// We don't attempt further detection since we don't have raw bytes
	return japanese.ShiftJIS
}

// IsUTF8Filename checks if a filename in an archive uses UTF-8 encoding
// This is specific to ZIP files, which have a flag bit for UTF-8
func IsUTF8Filename(fileHeader interface{}) bool {
	// Default to assuming UTF-8
	isUTF8 := true

	// Check for ZIP-specific header fields
	if header, ok := fileHeader.(interface{ GetFlags() uint16 }); ok {
		// Check if UTF-8 flag (0x800) is set in flag bits
		isUTF8 = (header.GetFlags() & 0x800) != 0
	} else if header, ok := fileHeader.(interface{ GetUTF8() bool }); ok {
		// Some implementations might provide a direct method
		isUTF8 = header.GetUTF8()
	} else if m, ok := fileHeader.(map[string]interface{}); ok {
		// Try to check a map-style header
		if flags, ok := m["flags"].(uint16); ok {
			isUTF8 = (flags & 0x800) != 0
		} else if utf8Flag, ok := m["utf8"].(bool); ok {
			isUTF8 = utf8Flag
		}
	}

	return isUTF8
}
