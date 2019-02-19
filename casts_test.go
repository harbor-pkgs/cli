package cli_test

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/harbor-pkgs/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestInvalidStoreType(t *testing.T) {
	var integer64 int64
	var integer int
	var aInt [2]int
	var aStr [2]string
	var aBool [2]bool

	tests := []struct {
		option *cli.Option
		err    string
	}{
		{
			option: &cli.Option{Name: "foo", Store: &integer64},
			err:    "invalid 'Store' while adding option 'foo': cannot store 'int64'; type not supported",
		},
		{
			option: &cli.Option{Name: "foo", Store: integer},
			err:    "invalid 'Store' while adding option 'foo': cannot use non pointer type 'int'; must provide a pointer",
		},
		{
			option: &cli.Option{Name: "foo", Store: &aInt},
			err:    "invalid 'Store' while adding option 'foo': cannot store array of type int; only slices supported",
		},
		{
			option: &cli.Option{Name: "foo", Store: &aStr},
			err:    "invalid 'Store' while adding option 'foo': cannot store array of type string; only slices supported",
		},
		{
			option: &cli.Option{Name: "foo", Store: &aBool},
			err:    "invalid 'Store' while adding option 'foo': cannot store array of type bool; only slices supported",
		},
	}

	for _, test := range tests {
		p := cli.New(nil)
		p.Add(test.option)

		retCode, err := p.Parse(nil, []string{})
		assert.NotNil(t, err)
		assert.Equal(t, cli.ErrorRetCode, retCode)
		assert.Contains(t, err.Error(), test.err)
	}
}

func TestMultpleOptionSameStore(t *testing.T) {
	var foo string
	p := cli.New(nil)
	p.Add(&cli.Option{Name: "foo", Store: &foo})
	p.Add(&cli.Option{Name: "bar", Store: &foo})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", "bang"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bang", foo)

	// Given
	retCode, err = p.Parse(nil, []string{"--bar", "foo"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "foo", foo)
}

func TestOptionDefaultScalar(t *testing.T) {
	var foo string
	p := cli.New(nil)
	p.Add(&cli.Option{Name: "foo", Store: &foo, Default: "bash"})

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

func TestOptionDefaultList(t *testing.T) {
	var foo []string
	p := cli.New(nil)
	p.Add(&cli.Option{Name: "foo", Store: &foo, Default: "bash,bar,foo"})

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

func TestOptionDefaultMap(t *testing.T) {
	var foo map[string]string
	var count int

	p := cli.New(nil)
	p.Add(&cli.Option{Name: "foo", Store: &foo, Default: "bar=foo,foo=bar"})

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

func TestOptionWithBoolSlice(t *testing.T) {
	var foo []bool
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Option{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

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
	assert.Equal(t, "invalid value for option 'foo': 'foo' is not a boolean", err.Error())
}

func TestOptionWithSlice(t *testing.T) {
	var foo []string
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Option{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", "bar,bang", "-f", "foo"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
	assert.Equal(t, []string{"bar,bang", "foo"}, foo)
}

func TestOptionStringMap(t *testing.T) {
	var foo map[string]string
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Option{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

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

func TestOptionIntMap(t *testing.T) {
	var foo map[string]int
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Option{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

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
	assert.Equal(t, "invalid value for option 'foo': 'one' is not an integer", err.Error())
}

func TestOptionBoolMap(t *testing.T) {
	var foo map[string]bool
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Option{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

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
	assert.Equal(t, "invalid value for option 'foo': 'one' is not a boolean", err.Error())
}

func TestInvalidMapType(t *testing.T) {
	var foo map[string]int64
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Option{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", "bar=1"})

	// Then
	require.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Contains(t, err.Error(), "invalid 'Store' while adding option 'foo': cannot use 'map[string]int64';")
}

func TestOptionWithMapAndJSON(t *testing.T) {
	var foo map[string]string
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Option{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

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

type cordinates struct {
	points []point
}

type point struct {
	x int
	y int
}

func (c *cordinates) Set(v string) error {
	parts := strings.Split(v, ",")
	if len(parts) != 2 {
		return errors.New("malformed coordinate point")
	}

	var points []int
	for _, part := range parts {
		v, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return err
		}
		points = append(points, int(v))
	}

	c.points = append(c.points, point{x: points[0], y: points[1]})
	spew.Dump(c.points)
	return nil
}

func TestSetValueInterface(t *testing.T) {
	var cords cordinates

	p := cli.New(nil)
	p.Add(&cli.Option{
		Flags:   cli.CanRepeat | cli.NoSplit,
		Aliases: []string{"p"},
		Name:    "point",
		Store:   &cords,
	})

	// Given
	retCode, err := p.Parse(nil, []string{"--point", "1"})

	// Then
	assert.NotNil(t, err)
	assert.Equal(t, "invalid value for option 'point': malformed coordinate point", err.Error())

	// Given
	retCode, err = p.Parse(nil, []string{"--point", "1,2"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, []point{
		{x: 1, y: 2},
	}, cords.points)
	cords.points = nil

	// Given
	retCode, err = p.Parse(nil, []string{"--point", "1,2", "-p", "25,35", "-p", "100,5000"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, []point{
		{x: 1, y: 2},
		{x: 25, y: 35},
		{x: 100, y: 5000},
	}, cords.points)

}

// TODO: func TestMapKind(t *testing.T) {
// TODO: func TestSliceKind(t *testing.T) {

type TestStruct struct {
	StringOpt   string
	IntOpt      int
	Int64Opt    int64
	Uint64Opt   uint64
	UintOpt     uint
	Float64Opt  float64
	BoolOpt     bool
	DurationOpt time.Duration
}

func TestBoolConv(t *testing.T) {
	var ts TestStruct

	tests := []struct {
		v   cli.Variant
		exp bool
		val string
	}{
		{v: &cli.Option{Name: "foo", Store: &ts.BoolOpt}, val: "true", exp: true},
		{v: &cli.Option{Name: "foo", Store: &ts.BoolOpt}, val: "TRUE", exp: true},
		{v: &cli.Option{Name: "foo", Store: &ts.BoolOpt}, val: "True", exp: true},
		{v: &cli.Option{Name: "foo", Store: &ts.BoolOpt}, val: "false", exp: false},
		{v: &cli.Option{Name: "foo", Store: &ts.BoolOpt}, val: "FALSE", exp: false},
		{v: &cli.Option{Name: "foo", Store: &ts.BoolOpt}, val: "False", exp: false},
		{v: &cli.Option{Name: "foo", Store: &ts.BoolOpt}, val: "yes", exp: true},
		{v: &cli.Option{Name: "foo", Store: &ts.BoolOpt}, val: "YES", exp: true},
		{v: &cli.Option{Name: "foo", Store: &ts.BoolOpt}, val: "Yes", exp: true},
		{v: &cli.Option{Name: "foo", Store: &ts.BoolOpt}, val: "no", exp: false},
		{v: &cli.Option{Name: "foo", Store: &ts.BoolOpt}, val: "NO", exp: false},
		{v: &cli.Option{Name: "foo", Store: &ts.BoolOpt}, val: "No", exp: false},
		{v: &cli.Option{Name: "foo", Store: &ts.BoolOpt}, val: "0", exp: false},
		{v: &cli.Option{Name: "foo", Store: &ts.BoolOpt}, val: "1", exp: true},
	}

	for i, test := range tests {
		ts = TestStruct{}
		p := cli.New(nil)
		p.Add(test.v)
		retCode, err := p.Parse(nil, []string{"--foo", test.val})

		// Then
		assert.Nil(t, err)
		assert.Equal(t, 0, retCode)
		assert.Equal(t, ts.BoolOpt, test.exp, fmt.Sprintf("test case '%d'", i))
	}
}

func TestScalarKind(t *testing.T) {
	var ts TestStruct

	tests := []struct {
		v    cli.Variant
		cmp  func(string)
		args []string
		env  string
	}{

		// String
		{
			cmp:  func(msg string) { assert.Equal(t, "foobar", ts.StringOpt, msg) },
			v:    &cli.Option{Name: "foo", Store: &ts.StringOpt},
			args: []string{"--foo", "foobar"},
		},
		{
			cmp:  func(msg string) { assert.Equal(t, "default-foo", ts.StringOpt, msg) },
			v:    &cli.Option{Name: "foo", Store: &ts.StringOpt, Default: "default-foo"},
			args: []string{},
		},
		{
			cmp:  func(msg string) { assert.Equal(t, "env-foo", ts.StringOpt, msg) },
			v:    &cli.Option{Name: "foo", Store: &ts.StringOpt, Env: "FOO"},
			args: []string{},
			env:  "env-foo",
		},

		// Integer
		{
			cmp:  func(msg string) { assert.Equal(t, 42, ts.IntOpt, msg) },
			v:    &cli.Option{Name: "foo", Store: &ts.IntOpt},
			args: []string{"--foo", "42"},
		},
		{
			cmp:  func(msg string) { assert.Equal(t, 255, ts.IntOpt, msg) },
			v:    &cli.Option{Name: "foo", Store: &ts.IntOpt, Default: "255"},
			args: []string{},
		},
		{
			cmp:  func(msg string) { assert.Equal(t, 500, ts.IntOpt, msg) },
			v:    &cli.Option{Name: "foo", Store: &ts.IntOpt, Env: "FOO"},
			args: []string{},
			env:  "500",
		},

		// Boolean
		{
			cmp:  func(msg string) { assert.Equal(t, true, ts.BoolOpt, msg) },
			v:    &cli.Option{Name: "foo", Store: &ts.BoolOpt},
			args: []string{"--foo", "true"},
		},
		{
			cmp:  func(msg string) { assert.Equal(t, true, ts.BoolOpt, msg) },
			v:    &cli.Option{Name: "foo", Store: &ts.BoolOpt, Default: "true"},
			args: []string{},
		},
		{
			cmp:  func(msg string) { assert.Equal(t, true, ts.BoolOpt, msg) },
			v:    &cli.Option{Name: "foo", Store: &ts.BoolOpt, Env: "FOO"},
			args: []string{},
			env:  "true",
		},
	}

	for i, test := range tests {
		testCase := fmt.Sprintf("test case '%d'", i)
		os.Setenv("FOO", test.env)
		ts = TestStruct{}

		p := cli.New(nil)
		p.Add(test.v)
		retCode, err := p.Parse(nil, test.args)

		assert.Nil(t, err, testCase)
		assert.Equal(t, 0, retCode, testCase)
		test.cmp(testCase)
	}

	/* TODO &cli.Option{Name: "int64", Store: &int64Opt},
		&cli.Option{Name: "uint", Store: &uintOpt},
		&cli.Option{Name: "uint64", Store: &uint64Opt},
		&cli.Option{Name: "float64", Store: &float64Opt},
		&cli.Option{Name: "duration", Store: &durationOpt},

	// Given no value
	retCode, err := p.Parse(nil, []string{
		"--bool", "yes",
		"--int", "42",
		"--int64", "500",
		"--uint", "0xC0",
		"--uint64", "0xC0FFE",
		"--string", "foobar",
		"--float64", "3.14",
		"--duration", "2m",
	})*/
}

// TODO: Test all possible type conversions
// TODO: Add all the types that std 'flag' package supports
