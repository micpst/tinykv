package hash

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"sort"
)

type VolumeMetric struct {
	Score  []byte
	Volume string
}

type VolumeMetrics []VolumeMetric

func (m VolumeMetrics) Len() int { return len(m) }

func (m VolumeMetrics) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

func (m VolumeMetrics) Less(i, j int) bool {
	return bytes.Compare(m[i].Score, m[j].Score) == 1
}

func KeyToPath(key []byte) string {
	md5Key := md5.Sum(key)
	b64Key := base64.StdEncoding.EncodeToString(key)
	return fmt.Sprintf("/%02x/%02x/%s", md5Key[0], md5Key[1], b64Key)
}

func KeyToVolumes(key []byte, volumes []string, replicas int) []string {
	selected := make([]string, replicas)
	metrics := make(VolumeMetrics, replicas)

	for _, volume := range volumes {
		hash := md5.New()
		hash.Write(key)
		hash.Write([]byte(volume))
		score := hash.Sum(nil)
		metrics = append(metrics, VolumeMetric{score, volume})
	}

	sort.Stable(metrics)

	for i := 0; i < replicas; i++ {
		selected[i] = metrics[i].Volume
	}

	return selected
}
