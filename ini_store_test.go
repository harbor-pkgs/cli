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
	kv, err := cli.NewIniStore(bytes.NewReader([]byte(scalarFile)))
	require.Nil(t, err)

	value, count, err := kv.Get(context.TODO(), "foo", cli.ScalarKind)
	require.Nil(t, err)
	assert.Equal(t, 4, count)
	assert.Equal(t, "bar", value.(string))
}

func TestReadListValues(t *testing.T) {
	kv, err := cli.NewIniStore(bytes.NewReader([]byte(scalarFile)))
	require.Nil(t, err)

	value, count, err := kv.Get(context.TODO(), "foo", cli.ListKind)
	require.Nil(t, err)
	assert.Equal(t, 4, count)

	list := value.([]string)
	sort.Strings(list)

	assert.Equal(t, []string{"", "bang", "bar", "bash"}, list)
}

func TestReadMapValues(t *testing.T) {
	kv, err := cli.NewIniStore(bytes.NewReader([]byte(mapFile)))
	require.Nil(t, err)

	value, count, err := kv.Get(context.TODO(), "foo", cli.MapKind)
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
	kv, err := cli.NewIniStore(bytes.NewReader([]byte(mapFile)))
	require.Nil(t, err)

	value, count, err := kv.Get(context.TODO(), "key", cli.MapKind)
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

	kv, err := cli.NewIniStore(bytes.NewReader([]byte(scalarFile)))
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

var castFile = `
# Strings
str=bar
strSlice=bar,foo,bang
strMap="bash=bang,bar=foo,bang=bash"

# Integers
integer=1
intSlice=1,2,3
intMap="one=1,two=2,three=3"

# Booleans
boolean=true
boolSlice=false,true,false
boolMap="on=true,off=false,yes=no"
`

func TestCastFile(t *testing.T) {
	var str string
	var strSlice []string
	var strMap map[string]string
	var integer int
	var intSlice []int
	var intMap map[string]int
	var boolean bool
	var boolSlice []bool
	var boolMap map[string]bool

	p := cli.New(nil)
	p.Add(
		&cli.Flag{Name: "str", Store: &str},
		&cli.Flag{Name: "strSlice", Store: &strSlice},
		&cli.Flag{Name: "strMap", Store: &strMap},

		&cli.Flag{Name: "integer", Store: &integer},
		&cli.Flag{Name: "intSlice", Store: &intSlice},
		&cli.Flag{Name: "intMap", Store: &intMap},

		&cli.Flag{Name: "boolean", Store: &boolean},
		&cli.Flag{Name: "boolSlice", Store: &boolSlice},
		&cli.Flag{Name: "boolMap", Store: &boolMap},
	)
	kv, err := cli.NewIniStore(bytes.NewReader([]byte(castFile)))
	require.Nil(t, err)
	p.AddStore(kv)

	// Given no value
	retCode, err := p.Parse(nil, []string{})

	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar", str)
	assert.Equal(t, []string{"bar", "foo", "bang"}, strSlice)
	assert.Equal(t, map[string]string{"bash": "bang", "bar": "foo", "bang": "bash"}, strMap)

	assert.Equal(t, 1, integer)
	assert.Equal(t, []int{1, 2, 3}, intSlice)
	assert.Equal(t, map[string]int{"one": 1, "two": 2, "three": 3}, intMap)

	assert.Equal(t, true, boolean)
	assert.Equal(t, []bool{false, true, false}, boolSlice)
	assert.Equal(t, map[string]bool{"on": true, "off": false, "yes": false}, boolMap)
}
