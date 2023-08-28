package discordutil

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"time"
)

// Returns 32 byte string
func NewNonce() (nonce string, err error) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	randomBytes := make([]byte, 8)
	nonceBytes := make([]byte, 16)

	_, err = rand.Read(randomBytes)

	if err != nil {
		return
	}

	copy(nonceBytes, randomBytes)
	binary.BigEndian.PutUint64(nonceBytes[8:], uint64(timestamp))

	nonce = hex.EncodeToString(nonceBytes)

	return
}
