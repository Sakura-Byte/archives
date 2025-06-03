package archives

import (
	"github.com/saintfish/chardet"
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

// DetectEncoding analyzes the provided byte data to determine its encoding
func DetectEncoding(data []byte) (encoding.Encoding, error) {
	if len(data) == 0 {
		return nil, nil
	}

	// Use chardet for encoding detection
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(data)
	if err != nil {
		return nil, err
	}

	// Convert the detected charset to an encoding.Encoding
	return GetEncodingFromCharset(result.Charset, result.Language), nil
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
