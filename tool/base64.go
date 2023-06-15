package tool

import "encoding/base64"

// default base64
func Base64Encode(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// url base64
func Base64URLEncode(input string) string {
	return base64.URLEncoding.EncodeToString([]byte(input))
}

// default base64
func Base64Decode(input string) string {
	if r, err := base64.StdEncoding.DecodeString(input); err != nil {
		return ""
	} else {
		return string(r)
	}
}

// url base64
func Base64URLDecode(input string) string {
	if r, err := base64.URLEncoding.DecodeString(input); err != nil {
		return ""
	} else {
		return string(r)
	}
}