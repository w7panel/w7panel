package longhorn

import (
	"testing"
)

func TestContainsAll(t *testing.T) {
	testCases := []struct {
		a        []string
		b        []string
		expected bool
	}{
		{
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c"},
			expected: true,
		},
		{
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b"},
			expected: true,
		},
		{
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "d"},
			expected: false,
		},
		{
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c", "d"},
			expected: false,
		},
		{
			a:        []string{},
			b:        []string{},
			expected: true,
		},
		{
			a:        []string{},
			b:        []string{"a", "b", "c"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		result := containsAll(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("Expected %v, but got %v for a=%v and b=%v", tc.expected, result, tc.a, tc.b)
		}
	}
}

func TestCreateVolume(t *testing.T) {
	LonghorStoragePercentage("900")
}
