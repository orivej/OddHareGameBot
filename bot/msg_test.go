package bot

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestNum(t *testing.T) {
	for expected, actual := range map[string]string{
		"0 слов":  num(0, "слов", "слово", "слова"),
		"1 слово": num(1, "слов", "слово", "слова"),
		"2 слова": num(2, "слов", "слово", "слова"),
		"4 слова": num(4, "слов", "слово", "слова"),
		"5 слов":  num(5, "слов", "слово", "слова"),
	} {
		assert.Equal(t, expected, actual)
	}
}

func TestJoinWithAnd(t *testing.T) {
	for _, xs := range [][]string{
		{""},
		{"x", "x"},
		{"x и y", "x", "y"},
		{"x, y и z", "x", "y", "z"},
	} {
		assert.Equal(t, xs[0], joinWithAnd(xs[1:]))
	}
}

func TestJoinEnumerate(t *testing.T) {
	for _, xs := range [][]string{
		{""},
		{"сначала x", "x"},
		{"сначала x, затем y", "x", "y"},
		{"сначала x, затем y и z", "x", "y", "z"},
		{"сначала x, затем y, z и ~", "x", "y", "z", "~"},
	} {
		assert.Equal(t, xs[0], joinEnumerate(xs[1:]))
	}
}
