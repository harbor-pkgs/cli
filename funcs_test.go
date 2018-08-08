package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

func TestStringToMapMalformed(t *testing.T) {
	_, err := cli.StringToMap("foo")
	require.NotNil(t, err)
	assert.Equal(t, "expected '=' after 'foo' but found ''; map values should be 'key=value' separated by commas", err.Error())

	_, err = cli.StringToMap(",")
	require.NotNil(t, err)
	assert.Equal(t, "expected '=' after ',' but found ''; map values should be 'key=value' separated by commas", err.Error())

	_, err = cli.StringToMap("=")
	require.NotNil(t, err)
	assert.Equal(t, "expected '=' after '=' but found ''; map values should be 'key=value' separated by commas", err.Error())

	_, err = cli.StringToMap("=,")
	require.NotNil(t, err)
	assert.Equal(t, "expected '=' after '=' but found ','; map values should be 'key=value' separated by commas", err.Error())
}

func TestStringToMapWithJSON(t *testing.T) {
	strMap, err := cli.StringToMap(`{"belt":"car","table":"cloth"}`)
	require.Nil(t, err)
	assert.Contains(t, strMap, "belt")
	assert.Contains(t, strMap, "table")
	assert.Equal(t, strMap["belt"], "car")
	assert.Equal(t, strMap["table"], "cloth")
}

func TestStringToMapWithEmptyString(t *testing.T) {
	strMap, err := cli.StringToMap("")
	require.NotNil(t, err)
	require.Equal(t, len(strMap), 0)
	require.Equal(t, "expected key at pos '0' but found none; map values should be 'key=value' separated by commas", err.Error())
}

func TestStringToMapWithNoValue(t *testing.T) {
	strMap, err := cli.StringToMap("foo=")
	require.NotNil(t, err)
	require.Equal(t, len(strMap), 0)
	require.Equal(t, "expected value after '=' but found none; map values should be 'key=value' separated by commas", err.Error())
}

func TestStringToSlice(t *testing.T) {
	r := cli.StringToSlice("one,two,three")
	assert.Equal(t, []string{"one", "two", "three"}, r)

	r = cli.StringToSlice("one, two, three", strings.TrimSpace)
	assert.Equal(t, []string{"one", "two", "three"}, r)

	r = cli.StringToSlice("one, two, three", strings.TrimSpace, strings.ToUpper)
	assert.Equal(t, []string{"ONE", "TWO", "THREE"}, r)

	r = cli.StringToSlice("one ", strings.TrimSpace, strings.ToUpper)
	assert.Equal(t, []string{"ONE"}, r)
}

func ExampleCurlString() {
	// Payload
	payload, err := json.Marshal(map[string]string{
		"stuff": "junk",
	})

	// Create the new Request
	req, err := http.NewRequest("POST", "http://google.com/stuff", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", cli.CurlString(req, &payload))

	// Output:
	// curl -i -X POST http://google.com/stuff  -d '{"stuff":"junk"}'
}

func ExampleDedent() {
	desc := cli.Dedent(`Example is a fast and flexible thingy

	Complete documentation is available at http://thingy.com

	Example Usage:
	    $ example-cli some-argument
	    Hello World!`)

	fmt.Println(desc)
	// Output:
	// Example is a fast and flexible thingy
	//
	// Complete documentation is available at http://thingy.com
	//
	// Example Usage:
	//     $ example-cli some-argument
	//     Hello World!
}

func ExampleWordWrap() {
	msg := cli.WordWrap(`No code is the best way to write secure and reliable applications.
		Write nothing; deploy nowhere. This is just an example application, but imagine it doing 
		anything you want.`,
		3, 80)
	fmt.Println(msg)
	// Output:
	// No code is the best way to write secure and reliable applications.Write
	//    nothing; deploy nowhere. This is just an example application, but imagine it
	//    doing anything you want.
}

func ExampleStringToSlice() {
	// Returns []string{"one"}
	fmt.Println(cli.StringToSlice("one"))

	// Returns []string{"one", "two", "three"}
	fmt.Println(cli.StringToSlice("one, two, three", strings.TrimSpace))

	//  Returns []string{"ONE", "TWO", "THREE"}
	fmt.Println(cli.StringToSlice("one,two,three", strings.ToUpper, strings.TrimSpace))

	// Output:
	// [one]
	// [one two three]
	// [ONE TWO THREE]
}

func ExampleStringToMap() {
	// Returns map[string]string{"foo": "bar"}
	fmt.Println(cli.StringToMap("foo=bar"))

	// Returns map[string]string{"foo": "bar", "kit": "kitty kat"}
	m, _ := cli.StringToMap(`foo=bar,kit="kitty kat"`)
	fmt.Printf("foo: %s\n", m["foo"])
	fmt.Printf("kit: %s\n", m["kit"])

	// Returns map[string]string{"foo": "bar", "kit": "kitty kat"}
	m, _ = cli.StringToMap(`{"foo":"bar","kit":"kitty kat"}`)
	fmt.Printf("foo: %s\n", m["foo"])
	fmt.Printf("kit: %s\n", m["kit"])

	// Output:
	// map[foo:bar] <nil>
	// foo: bar
	// kit: "kitty kat"
	// foo: bar
	// kit: kitty kat
}
