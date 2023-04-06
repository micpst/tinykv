package hash

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
)

func KeyToPath(key []byte) string {
	mdKey := md5.Sum(key)
	b64Key := base64.StdEncoding.EncodeToString(key)
	return fmt.Sprintf("/%02x/%02x/%s", mdKey[0], mdKey[1], b64Key)
}

func KeyToVolume(key []byte, volumes []string) string {
	var selectedVolume string
	var bestScore []byte

	for _, volume := range volumes {
		hash := md5.New()
		hash.Write([]byte(volume))
		hash.Write(key)
		score := hash.Sum(nil)
		if bestScore == nil || bytes.Compare(bestScore, score) == -1 {
			bestScore = score
			selectedVolume = volume
		}
	}
	return selectedVolume
}
