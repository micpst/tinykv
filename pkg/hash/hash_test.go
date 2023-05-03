package hash

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCase[Given any, Expected any] struct {
	given    Given
	expected Expected
}

type KeyToVolumesInput struct {
	key      []byte
	volumes  []string
	replicas int
}

func TestKeyToPath(t *testing.T) {
	cases := []TestCase[[]byte, string]{
		{
			given:    []byte("hello"),
			expected: "/5d/41/aGVsbG8=",
		},
		{
			given:    []byte("bye"),
			expected: "/bf/a9/Ynll",
		},
		{
			given:    []byte("weeeeeeeee"),
			expected: "/8f/a6/d2VlZWVlZWVlZQ==",
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			actual := KeyToPath(c.given)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestKeyToVolumes(t *testing.T) {
	cases := []TestCase[KeyToVolumesInput, []string]{
		{
			given: KeyToVolumesInput{
				key: []byte("hello"),
				volumes: []string{
					"10.23.0.1:80",
					"10.23.0.2:80",
					"10.23.0.3:80",
					"10.23.0.4:80",
				},
				replicas: 1,
			},
			expected: []string{
				"10.23.0.3:80",
			},
		},
		{
			given: KeyToVolumesInput{
				key: []byte("hello"),
				volumes: []string{
					"10.23.0.1:80",
					"10.23.0.2:80",
					"10.23.0.3:80",
					"10.23.0.4:80",
				},
				replicas: 3,
			},
			expected: []string{
				"10.23.0.3:80",
				"10.23.0.1:80",
				"10.23.0.4:80",
			},
		},
		{
			given: KeyToVolumesInput{
				key: []byte("hello"),
				volumes: []string{
					"10.23.0.1:80",
					"10.23.0.2:80",
					"10.23.0.3:80",
					"10.23.0.4:80",
				},
				replicas: 5,
			},
			expected: []string{
				"10.23.0.3:80",
				"10.23.0.1:80",
				"10.23.0.4:80",
				"10.23.0.2:80",
				"",
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			actual := KeyToVolumes(c.given.key, c.given.volumes, c.given.replicas)
			assert.Equal(t, c.expected, actual)
		})
	}
}
