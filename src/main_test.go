package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test1(t *testing.T) {
	expected := true
	actual := true
	require.Equal(t, actual, expected)
}
