package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func TestI18nLanguageTag(t *testing.T) {
	tag, err := language.Parse("zh-CN")
	require.NoError(t, err)
	assert.Equal(t, "zh-CN", tag.String())
}
