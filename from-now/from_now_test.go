package fromNow

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/spf13/cobra"
	assert "github.com/stretchr/testify/require"
)

var (
	// TIMESTAMP_REGEXP is the regular expression that all timestamps should conform to
	// we assume we won't need to validate beyond 2099, this is only used for test purposes
	TIMESTAMP_REGEXP = regexp.MustCompile("^20[0-9]{2}-(1[0-2]|0[0-9])-(3[0-1]|[0-2][0-9])T(2[0-3]|[01][0-9])(:[0-5][0-9]){2}(Z|(\\+|-)(2[0-3]|[01][0-9]):[0-5][0-9])$")
)

func TestParseTimeComplete(t *testing.T) {
	assert := assert.New(t)
	offset, err := parseTime("1 year 2 months 3 weeks 4 days 5 hours 6 minutes 7 seconds")

	assert.NoError(err, "the string should parse without error")
	assert.Equal(offset.years, 1)
	assert.Equal(offset.months, 2)
	assert.Equal(offset.weeks, 3)
	assert.Equal(offset.days, 4)
	assert.Equal(offset.hours, 5)
	assert.Equal(offset.minutes, 6)
	assert.Equal(offset.seconds, 7)
}

func TestParseTimeIncomplete(t *testing.T) {
	assert := assert.New(t)

	// Test if we omit some fields
	offset, err := parseTime("2 years 5 days 6 minutes")

	assert.NoError(err, "the string should parse without error")
	assert.Equal(offset.years, 2)
	assert.Equal(offset.months, 0)
	assert.Equal(offset.weeks, 0)
	assert.Equal(offset.days, 5)
	assert.Equal(offset.hours, 0)
	assert.Equal(offset.minutes, 6)
	assert.Equal(offset.seconds, 0)
}

func TestParseTimeInvalid(t *testing.T) {
	assert := assert.New(t)

	// Test if it's a valid time expression.
	_, err := parseTime("this should produce an error.")
	assert.Error(err, "the string should produce an error")
}

func TestAtoiHelper(t *testing.T) {
	assert := assert.New(t)

	assert.NotPanics(func() {
		atoiHelper("1")
	}, "should not panic")

	assert.Panics(func() {
		atoiHelper("!")
	}, "should panic")
}

// now, test the command as a whole

func TestFromNowInvalid(t *testing.T) {
	assert := assert.New(t)

	buf := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOutput(buf)

	err := fromNow(cmd, []string{"this should produce an error"})

	assert.Error(err, "the input is invalid and should produce an error")
}

func TestFromNowValid(t *testing.T) {
	assert := assert.New(t)

	buf := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOutput(buf)

	err := fromNow(cmd, []string{"2 years 5 days 6 minutes"})

	output := buf.String()
	output = output[0 : len(output)-1]
	match := TIMESTAMP_REGEXP.MatchString(output)

	assert.NoError(err, "error when given a valid input")
	assert.True(match, "the command did not return a valid timestamp")
}
