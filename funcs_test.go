package cli_test

import (
	"testing"

	"github.com/harbor-pkgs/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringToMapWithAlpha(t *testing.T) {
	strMap, err := cli.StringToMap("http.ip=192.168.1.1")
	require.Nil(t, err)
	assert.Contains(t, strMap, "http.ip")
	assert.Equal(t, strMap["http.ip"], "192.168.1.1")
}

func TestStringToMapWithEscape(t *testing.T) {
	strMap, err := cli.StringToMap(`http\=ip=192.168.1.1`)
	require.Nil(t, err)
	assert.Contains(t, strMap, "http=ip")
	assert.Equal(t, strMap["http=ip"], "192.168.1.1")
}

func TestStringToMapWithCommas(t *testing.T) {
	strMap, err := cli.StringToMap("foo=bar,bar=foo")
	require.Nil(t, err)
	assert.Contains(t, strMap, "foo")
	assert.Contains(t, strMap, "bar")
	assert.Equal(t, strMap["foo"], "bar")
	assert.Equal(t, strMap["bar"], "foo")
}

func TestStringToMapWithJSON(t *testing.T) {
	strMap, err := cli.StringToMap(`{"belt":"car","table":"cloth"}`)
	require.Nil(t, err)
	assert.Contains(t, strMap, "belt")
	assert.Contains(t, strMap, "table")
	assert.Equal(t, strMap["belt"], "car")
	assert.Equal(t, strMap["table"], "cloth")
}

func TestStringToMapWithNoValue(t *testing.T) {
	strMap, err := cli.StringToMap("")
	require.NotNil(t, err)
	require.Equal(t, len(strMap), 0)
	require.Equal(t, "expected key at pos '0' but found none; map values should be 'key=value' separated by commas", err.Error())
}
