package strgen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Pixelgaffer/dicod/strgen"
)

func TestChoice(t *testing.T) {
	channel, amount, err := strgen.GenerateStrings("\\(a|b)")
	assert.Nil(t, err)
	assert.Equal(t, int64(2), amount, "should produce two results")
	assert.EqualValues(t, []string{"a", "b"}, []string{<-channel, <-channel}, []string{"should be 'a'", "should be 'b'"})
}

func ExampleChoice() {
	c, _, _ := strgen.GenerateStrings("\\(foo|bar|baz)")
	for s := range c {
		fmt.Println(s)
	}
	// Output:
	// foo
	// bar
	// baz
}

func TestRangeAmount(t *testing.T) {
	assertAmount := func(s string, amount int64) {
		_, actual, err := strgen.GenerateStrings(s)
		assert.Nil(t, err)
		assert.Equal(t, amount, actual, fmt.Sprintf("%v should produce %v results", s, amount))
	}
	// Finite, Integer ranges
	assertAmount("\\[0..3]", 4)
	assertAmount("\\[2..4]", 3)
	assertAmount("\\[5..2]", 4)

	// Finite, FP ranges
	assertAmount("\\[1.5..2]", 1)
	assertAmount("\\[1.5..3]", 2)

	// Finite, Integer ranges w/ step
	assertAmount("\\[0..2..3]", 2)
	assertAmount("\\[-2..0.5..0]", 5)
	assertAmount("\\[5..-1..2]", 4)

	// Infinite
	assertAmount("\\[0..]", -1)
	assertAmount("\\[-42..]", -1)
}

func ExampleRange() {
	c, _, _ := strgen.GenerateStrings("\\[0..0.5..2]")
	for s := range c {
		fmt.Println(s)
	}
	// Output:
	// 0
	// 0.5
	// 1
	// 1.5
	// 2
}

func TestInvalidInput(t *testing.T) {
	_, _, err := strgen.GenerateStrings("\\[0..bar.\\]foo")
	assert.NotNil(t, err)
	_, _, err = strgen.GenerateStrings("\\[5..1..0]")
	assert.NotNil(t, err)
}

func TestBasicText(t *testing.T) {
	channel, amount, err := strgen.GenerateStrings("foo bar")
	assert.Nil(t, err)
	assert.Equal(t, int64(1), amount, "should produce one result")
	assert.Equal(t, "foo bar", <-channel, "should produce the correct result")
}
