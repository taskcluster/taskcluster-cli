package slugid

import (
	"fmt"
	//"os"
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
)

func init() {
	// the basic slugid command
	slugid := &cobra.Command{
		Use:   "slugid",
		Short: "Generates V4 UUIDs and encodes/decodes them from/to 22 character URL-safe base64 slugs.",
	}
	root.Command.AddCommand(slugid)

	// v4
	slugid.AddCommand(&cobra.Command{
		Use:   "v4",
		Short: "Generates the slug of a V4 UUID.",
		Run:   generateV4,
	})

	// nice
	slugid.AddCommand(&cobra.Command{
		Use:   "nice",
		Short: "Generates the slug of a V4 UUID in a 'nice' format.",
		Run:   generateNice,
	})

	// decode
	slugid.AddCommand(&cobra.Command{
		Use:   "decode",
		Short: "Decodes a slug into a UUID.",
		RunE:  decode,
	})

	// encode
	slugid.AddCommand(&cobra.Command{
		Use:   "encode",
		Short: "Encode an UUID into a slug.",
		RunE:  encode,
	})
}

// generateV4 generates a normal v4 uuid
func generateV4(_ *cobra.Command, _ []string) {
	fmt.Println(sluglib.V4())
}

// generateNice generates uuid with "nice" properties
func generateNice(_ *cobra.Command, _ []string) {
	fmt.Println(sluglib.Nice())
}

// decode decodes a slug into a uuid
func decode(_ *cobra.Command, args []string) error {
	slug := args[0]

	// nice slugs are just a subset of all slugs, which must match V4 pattern
	// this slug may be nice or not; we don't know, so use general pattern
	match := RegexpSlugV4.MatchString(slug)
	if !match {
		return fmt.Errorf("invalid slug format '%s'", slug)
	}

	// and decode
	fmt.Println(sluglib.Decode(slug))
	return nil
}

// encodes uuid into a slug
func encode(_ *cobra.Command, args []string) error {
	uuid := args[0]

	// nice slugs are just a subset of all slugs, which must match V4 pattern
	// this slug may be nice or not; we don't know, so use general pattern
	match := RegexpUUIDV4.MatchString(uuid)
	if match == false {
		return fmt.Errorf("invalid uuid format '%s'", uuid)
	}

	// the uuid string needs to be parsed into uuidlib.UUID before encoding
	fmt.Println(sluglib.Encode(uuidlib.Parse(uuid)))
	return nil
}
