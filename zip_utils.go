package archives

import (
	"bytes"

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
	// Note: chardet.Detector doesn't have an EnabledDetectors field, so we use default detectors

	result, err := detector.DetectBest(data)
	if err != nil {
		return nil, err
	}

	// Log detection results for debugging (comment out in production)
	// fmt.Printf("Detected encoding: %s, language: %s, confidence: %.2f%%\n",
	//           result.Charset, result.Language, result.Confidence*100)

	// If confidence is too low, try more aggressive detection
	if float64(result.Confidence) < 0.7 {
		// Try to identify based on specific byte patterns
		if containsJapaneseBytes(data) {
			return japanese.ShiftJIS, nil
		} else if containsKoreanBytes(data) {
			return korean.EUCKR, nil
		} else if containsChineseBytes(data) {
			return simplifiedchinese.GBK, nil
		}
	}

	// Convert the detected charset to an encoding.Encoding
	enc := GetEncodingFromCharset(result.Charset, result.Language)
	if enc != nil {
		return enc, nil
	}

	// If we couldn't determine the encoding explicitly, try the fallbacks
	for _, enc := range GetFallbackEncodings() {
		// Try to decode a sample with this encoding
		decoder := enc.NewDecoder()
		_, err := decoder.Bytes(data)
		if err == nil {
			return enc, nil
		}
	}

	// Default to ShiftJIS as most common problematic encoding in ZIP files
	return japanese.ShiftJIS, nil
}

// containsJapaneseBytes checks for byte patterns common in Japanese encodings
func containsJapaneseBytes(data []byte) bool {
	// Common byte patterns in Shift-JIS
	return bytes.Contains(data, []byte{0x82, 0xA0}) || // Hiragana markers
		bytes.Contains(data, []byte{0x83, 0x40}) || // Katakana markers
		bytes.Contains(data, []byte{0x82, 0x6A}) || // Kanji range markers
		bytes.Contains(data, []byte{0x8A, 0xBF}) // More Kanji markers
}

// containsKoreanBytes checks for byte patterns common in Korean encodings
func containsKoreanBytes(data []byte) bool {
	// Common byte patterns in EUC-KR
	return bytes.Contains(data, []byte{0xB0, 0xA1}) || // Hangul markers
		bytes.Contains(data, []byte{0xB0, 0xFA}) ||
		bytes.Contains(data, []byte{0xC7, 0xD1})
}

// containsChineseBytes checks for byte patterns common in Chinese encodings
func containsChineseBytes(data []byte) bool {
	// Common byte patterns in GBK
	return bytes.Contains(data, []byte{0xD6, 0xD0}) || // Common Chinese characters
		bytes.Contains(data, []byte{0xCE, 0xC4}) ||
		bytes.Contains(data, []byte{0xD7, 0xD6})
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
