package cli_test

import (
	"testing"
	"github.com/harbor-pkgs/cli"
	"github.com/stretchr/testify/assert"
	"sort"
	"github.com/stretchr/testify/require"
)

func TestInvalidStoreType(t *testing.T) {
	var foo int64

	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{})

	// Then
	assert.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Contains(t, err.Error(), "invalid 'Store' while adding flag 'foo': cannot store 'int64'; type not supported")
}

// TODO: Add argument cast tests to the `TestFlag` tests and rename them `TestDefaultScalar()`, etc...
// TODO: Add tests that cast values from a KVStore, This will exercise the 'string'
// to 'type' path in `toXX` cast functions

func TestFlagDefaultScalar(t *testing.T) {
	var foo string
	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Default: "bash"})

	// Given no value
	retCode, err := p.Parse(nil, []string{})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bash", foo)

	// Given a value
	retCode, err = p.Parse(nil, []string{"--foo", "bar"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar", foo)
}

func TestFlagDefaultList(t *testing.T) {
	var foo []string
	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Default: "bash,bar,foo"})

	// Given no value
	retCode, err := p.Parse(nil, []string{})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	sort.Strings(foo)
	assert.Equal(t, []string{"bar", "bash", "foo"}, foo)

	// Given a value
	retCode, err = p.Parse(nil, []string{"--foo", "bar"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, []string{"bar"}, foo)
}

func TestFlagDefaultMap(t *testing.T) {
	var foo map[string]string
	var count int

	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Default: "bar=foo,foo=bar"})

	// Given
	retCode, err := p.Parse(nil, []string{})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 0, count)
	require.Contains(t, foo, "bar")
	require.Contains(t, foo, "foo")
	assert.Equal(t, foo["bar"], "foo")
	assert.Equal(t, foo["foo"], "bar")
}

func TestFlagWithBoolSlice(t *testing.T) {
	var foo []bool
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Positive test
	retCode, err := p.Parse(nil, []string{"--foo", "true", "-f", "false", "-f", "true"})
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 3, count)
	assert.Equal(t, []bool{true, false, true}, foo)

	// Negative test
	retCode, err = p.Parse(nil, []string{"--foo", "foo", "-f", "false", "-f", "true"})
	assert.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "invalid value for flag 'foo': 'foo' is not a boolean", err.Error())
}

func TestFlagWithSlice(t *testing.T) {
	var foo []string
	// TODO: Test array with 'Store' instead of slice 'var bar [2]string'
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", "bar,bang", "-f", "foo"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
	assert.Equal(t, []string{"bar,bang", "foo"}, foo)
}

func TestFlagStringMap(t *testing.T) {
	var foo map[string]string
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", "bar=foo", "-f", "foo=bar"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
	require.Contains(t, foo, "bar")
	require.Contains(t, foo, "foo")
	assert.Equal(t, foo["bar"], "foo")
	assert.Equal(t, foo["foo"], "bar")
}

func TestFlagIntMap(t *testing.T) {
	var foo map[string]int
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Positive test
	retCode, err := p.Parse(nil, []string{"--foo", "bar=1,cat=3", "-f", "foo=2"})
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
	require.Contains(t, foo, "bar")
	require.Contains(t, foo, "foo")
	require.Contains(t, foo, "cat")
	assert.Equal(t, foo["bar"], 1)
	assert.Equal(t, foo["foo"], 2)
	assert.Equal(t, foo["cat"], 3)

	// Negative test
	retCode, err = p.Parse(nil, []string{"--foo", "bar=one,cat=2", "-f", "foo=3"})
	require.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "invalid value for flag 'foo': 'one' is not an integer", err.Error())
}

func TestFlagBoolMap(t *testing.T) {
	var foo map[string]bool
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Positive test
	retCode, err := p.Parse(nil, []string{"--foo", "bar=true,cat=false", "-f", "foo=true"})
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
	require.Contains(t, foo, "bar")
	require.Contains(t, foo, "foo")
	require.Contains(t, foo, "cat")
	assert.Equal(t, foo["bar"], true)
	assert.Equal(t, foo["cat"], false)
	assert.Equal(t, foo["foo"], true)

	// Negative test
	retCode, err = p.Parse(nil, []string{"--foo", "bar=one,cat=false", "-f", "foo=true"})
	require.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "invalid value for flag 'foo': 'one' is not a boolean", err.Error())
}

func TestInvalidMapType(t *testing.T) {
	var foo map[string]int64
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", "bar=1"})

	// Then
	require.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Contains(t, err.Error(), "invalid 'Store' while adding flag 'foo': cannot use 'map[string]int64';")
}

func TestFlagWithMapAndJSON(t *testing.T) {
	var foo map[string]string
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", `{"bar":"foo"}`, "-f", `{"foo": "bar", "bash": "bang"}`})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
	require.Contains(t, foo, "bar")
	require.Contains(t, foo, "foo")
	require.Contains(t, foo, "bash")
	assert.Equal(t, foo["bar"], "foo")
	assert.Equal(t, foo["foo"], "bar")
	assert.Equal(t, foo["bash"], "bang")
}