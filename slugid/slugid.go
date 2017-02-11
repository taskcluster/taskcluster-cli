package slugid

import (
	"errors"
	"fmt"
	"regexp"

	uuidlib "github.com/pborman/uuid"
	"github.com/spf13/cobra"
	sluglib "github.com/taskcluster/slugid-go/slugid"

	"github.com/taskcluster/taskcluster-cli/root"
)

var (
	// See https://github.com/taskcluster/slugid-go/blob/master/README.md for
	// an explanation of these regular expressions. Note, compiling once is
	// more performant than compiling with each decode/encode call. We can use
	// regexp.MustCompile rather than regexp.Compile since these are constant
	// strings.

	// RegexpSlugV4 is the regular expression that all V4 Slug IDs should conform to
	RegexpSlugV4 = regexp.MustCompile("^[A-Za-z0-9_-]{8}[Q-T][A-Za-z0-9_-][CGKOSWaeimquy26-][A-Za-z0-9_-]{10}[AQgw]$")

	// RegexpUUIDV4 is the regular expression that all V4 UUIDs should conform to
	RegexpUUIDV4 = regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$")

	// RegexpSlugNice is the regular expression that all "nice" Slug IDs should conform to
	RegexpSlugNice = regexp.MustCompile("^[A-Za-f][A-Za-z0-9_-]{7}[Q-T][A-Za-z0-9_-][CGKOSWaeimquy26-][A-Za-z0-9_-]{10}[AQgw]$")

	// RegexpUUIDNice is the regular expression that all "nice" UUIDs should conform to
	RegexpUUIDNice = regexp.MustCompile("^[0-7][a-f0-9]{7}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$")

	// Command is the root of the slugid subtree.
	Command = &cobra.Command{
		Use:   "slugid",
		Short: "Generates V4 UUIDs and encodes/decodes them from/to 22 character URL-safe base64 slugs.",
	}
)

func init() {
	Command.AddCommand(
		// v4
		&cobra.Command{
			Use:   "v4",
			Short: "Generates a V4 UUID and output its slug.",
			Run:   printHelper(generateV4),
		},
		// nice
		&cobra.Command{
			Use:   "nice",
			Short: "Generates a 'nice' V4 UUID and output its slug.",
			Run:   printHelper(generateNice),
		},
		// decode
		&cobra.Command{
			Use:   "decode",
			Short: "Decodes a slug into a UUID.",
			RunE:  decode,
		},
		// encode
		&cobra.Command{
			Use:   "encode",
			Short: "Encode an UUID into a slug.",
			RunE:  encode,
		},
	)

	// Add the slugid subtree to the root.
	root.Command.AddCommand(Command)
}

// printHelper wraps simple functions and prints their result.
func printHelper(f func() string) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, _ []string) {
		fmt.Fprintln(cmd.OutOrStdout(), f())
	}
}

// generateV4 generates a normal v4 uuid
func generateV4() string {
	return sluglib.V4()
}

// generateNice generates a v4 uuid with "nice" properties
func generateNice() string {
	return sluglib.Nice()
}

// decode decodes a slug into a uuid
func decode(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("decode requires one argument")
	}
	slug := args[0]

	// nice slugs are just a subset of all slugs, which must match V4 pattern
	// this slug may be nice or not; we don't know, so use general pattern
	match := RegexpSlugV4.MatchString(slug)
	if match == false {
		return fmt.Errorf("invalid slug format '%s'", slug)
	}

	// and decode
	fmt.Fprintln(cmd.OutOrStdout(), sluglib.Decode(slug))
	return nil
}

// encode encodes a uuid into a slug
func encode(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("encode requires one argument")
	}
	uuid := args[0]

	// nice slugs are just a subset of all slugs, which must match V4 pattern
	// this slug may be nice or not; we don't know, so use general pattern
	match := RegexpUUIDV4.MatchString(uuid)
	if match == false {
		return fmt.Errorf("invalid uuid format '%s'", uuid)
	}

	// the uuid string needs to be parsed into uuidlib.UUID before encoding
	fmt.Fprintln(cmd.OutOrStdout(), sluglib.Encode(uuidlib.Parse(uuid)))
	return nil
}
