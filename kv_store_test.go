package cli_test

import (
	"bytes"
	"context"
	"sort"
	"testing"

	"github.com/harbor-pkgs/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var scalarFile = `
# a comment
foo=bar
"foo"="bang"
'foo'=bash
foo
bar=foo
`
var mapFile = `
# a comment
foo="bash=bang,bar=foo,alan=alan"
key='{"bash":"bang","bar":"foo","alan":"alan"}'
`

func TestReadScalarValues(t *testing.T) {
	kv, err := cli.NewKVStore(bytes.NewReader([]byte(scalarFile)))
	require.Nil(t, err)

	value, count, err := kv.Get(context.TODO(), "foo", cli.ScalarType)
	require.Nil(t, err)
	assert.Equal(t, 4, count)
	assert.Equal(t, "bar", value.(string))
}

func TestReadListValues(t *testing.T) {
	kv, err := cli.NewKVStore(bytes.NewReader([]byte(scalarFile)))
	require.Nil(t, err)

	value, count, err := kv.Get(context.TODO(), "foo", cli.ListType)
	require.Nil(t, err)
	assert.Equal(t, 4, count)

	list := value.([]string)
	sort.Strings(list)

	assert.Equal(t, []string{"", "bang", "bar", "bash"}, list)
}

func TestReadMapValues(t *testing.T) {
	kv, err := cli.NewKVStore(bytes.NewReader([]byte(mapFile)))
	require.Nil(t, err)

	value, count, err := kv.Get(context.TODO(), "foo", cli.MapType)
	require.Nil(t, err)
	assert.Equal(t, 1, count)

	values := value.(map[string]string)

	assert.Contains(t, values, "bash")
	assert.Contains(t, values, "bar")
	assert.Contains(t, values, "alan")
	assert.Equal(t, values["bash"], "bang")
	assert.Equal(t, values["bar"], "foo")
	assert.Equal(t, values["alan"], "alan")
}

func TestReadMapValuesJSON(t *testing.T) {
	kv, err := cli.NewKVStore(bytes.NewReader([]byte(mapFile)))
	require.Nil(t, err)

	value, count, err := kv.Get(context.TODO(), "key", cli.MapType)
	require.Nil(t, err)
	assert.Equal(t, 1, count)

	values := value.(map[string]string)

	assert.Contains(t, values, "bash")
	assert.Contains(t, values, "bar")
	assert.Contains(t, values, "alan")
	assert.Equal(t, values["bash"], "bang")
	assert.Equal(t, values["bar"], "foo")
	assert.Equal(t, values["alan"], "alan")
}

func TestFromStore(t *testing.T) {
	var foo, bar string
	var count int

	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo", Count: &count, Store: &foo})
	p.Add(&cli.Flag{Name: "bar", Count: &count, Store: &bar})

	kv, err := cli.NewKVStore(bytes.NewReader([]byte(scalarFile)))
	require.Nil(t, err)
	p.AddStore(kv)

	// Given no value
	retCode, err := p.Parse(nil, []string{"--foo", "thingy"})

	// Should prefer the argument provided on the
	// command line over the value from the store
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "thingy", foo)

	// But still provide values not specified on the command line
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "foo", bar)
}
