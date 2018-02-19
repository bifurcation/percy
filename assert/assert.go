package assert

import (
	"bytes"
	"fmt"
	"testing"
)

func True(t *testing.T, test bool, message string) {
	if !test {
		t.Fatal(message)
	}
}

func NotError(t *testing.T, err error, msg string) {
	True(t, err == nil, fmt.Sprintf("%s: %v", msg, err))
}

func Equal(t *testing.T, a, b interface{}, msg string) {
	True(t, a == b, fmt.Sprintf("%s: [%v] != [%v]", msg, a, b))
}

func NotEqual(t *testing.T, a, b interface{}, msg string) {
	True(t, a != b, fmt.Sprintf("%s: [%v] != [%v]", msg, a, b))
}

func BytesEqual(t *testing.T, a, b []byte, msg string) {
	True(t, bytes.Equal(a, b), fmt.Sprintf("%s: [%x] != [%x]", msg, a, b))
}

func BytesNotEqual(t *testing.T, a, b []byte, msg string) {
	True(t, !bytes.Equal(a, b), fmt.Sprintf("%s: [%x] =q= [%x]", msg, a, b))
}
