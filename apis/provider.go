//go:generate go run _codegen/fetch-apis.go

// Package apis implements all the API CommandProviders.
package apis

import (
	"bytes"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	got "github.com/taskcluster/go-got"
	"github.com/xeipuuv/gojsonschema"

	"github.com/taskcluster/taskcluster-cli/apis/definitions"
	"github.com/taskcluster/taskcluster-cli/client"
	"github.com/taskcluster/taskcluster-cli/cmds/root"
	"github.com/taskcluster/taskcluster-cli/config"
)

var (
	// Command is the root of the api commands.
	Command = &cobra.Command{
		Use:   "api",
		Short: "Direct access to TaskCluster APIs.",
	}
)

func init() {
	for name, service := range services {
		cmdString := strings.ToLower(name[0:1]) + name[1:]
		cmd := &cobra.Command{
			Use:   cmdString,
			Short: "Operates on the " + name + " service",
			Long:  service.Description,
		}
		for _, entry := range service.Entries {
			line := entry.Name
			for _, arg := range entry.Args {
				line += " <" + arg + ">"
			}

			subCmd := &cobra.Command{
				Use:   line,
				Short: entry.Title,
				Long:  buildHelp(&entry),
				RunE:  buildExecutor(entry, "api-"+name),
			}

			fs := subCmd.Flags()
			for _, q := range entry.Query {
				fs.String(q, "", "Specify the '"+q+"' query-string parameter")
			}

			cmd.AddCommand(subCmd)
		}

		fs := cmd.PersistentFlags()
		fs.StringP("base-url", "b", service.BaseURL, "BaseURL for "+cmdString)

		Command.AddCommand(cmd)
		config.RegisterOptions("api-"+name, map[string]config.OptionDefinition{
			"baseUrl": config.OptionDefinition{
				Default: service.BaseURL,
				Env:     "TASKCLUSTER_QUEUE_BASE_URL",
				Validate: func(value interface{}) error {
					return nil
				},
			},
		})
	}

	fs := Command.PersistentFlags()
	fs.StringP("output", "o", "-", "Output file")
	fs.BoolP("dry-run", "d", false, "Validate input against schema without making an actual request")
	Command.MarkPersistentFlagFilename("output")

	root.Command.AddCommand(Command)
}

func buildHelp(entry *definitions.Entry) string {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, "%s\n", entry.Title)
	fmt.Fprintf(buf, "Method:    %s\n", entry.Method)
	fmt.Fprintf(buf, "Path:      %s\n", entry.Route)
	fmt.Fprintf(buf, "Stability: %s\n", entry.Stability)
	fmt.Fprintf(buf, "Scopes:\n")
	for i, scopes := range entry.Scopes {
		fmt.Fprintf(buf, "  * %s", strings.Join(scopes, ","))
		if i < len(entry.Scopes)-1 {
			fmt.Fprintf(buf, ", or")
		}
		fmt.Fprintln(buf, "")
	}
	fmt.Fprintln(buf, "")
	fmt.Fprint(buf, entry.Description)

	return buf.String()
}

func buildExecutor(entry definitions.Entry, configKey string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Because cobra doesn't extract the args or map them to their
		// name, we build it ourselves.
		line := strings.Split(cmd.Use, " ")[1:]
		if len(args) < len(args) {
			return errors.New("Insufficient arguments given")
		}

		argmap := make(map[string]string)
		for i, a := range line {
			name := a[1 : len(a)-1]
			argmap[name] = args[i]
		}

		// Same with the local flags.
		query := make(map[string]string)
		fs := cmd.LocalFlags()
		for _, opt := range entry.Query {
			if val, err := fs.GetString(opt); err == nil {
				query[opt] = val
			} else {
				return err
			}
		}

		// Read payload if present
		var input io.Reader = os.Stdin
		if payload, ok := argmap["payload"]; ok {
			if payload != "-" {
				input = bytes.NewBufferString(payload)
			}
		}

		// Setup output
		var output = cmd.OutOrStdout()
		if flag := cmd.Flags().Lookup("output"); flag != nil && flag.Changed {
			filename := flag.Value.String()
			f, err := os.Create(filename)
			if err != nil {
				return fmt.Errorf("Failed to open output file, error: %s", err)
			}
			defer f.Close()
			output = f
		}

		if dry, _ := cmd.Flags().GetBool("dry-run"); dry {
			return validate(&entry, argmap, query, input, output)
		}

		baseURL := config.Configuration[configKey]["baseUrl"].(string)
		if flag := cmd.Flags().Lookup("base-url"); flag != nil && flag.Changed {
			baseURL = flag.Value.String()
		}

		return execute(baseURL, &entry, argmap, query, input, output)
	}
}

func validate(
	entry *definitions.Entry, args, query map[string]string,
	payload io.Reader, output io.Writer,
) error {
	// If there is no schema, there is nothing to validate
	schema, ok := schemas[entry.Input]
	if !ok {
		return nil
	}

	// Read all input
	data, err := ioutil.ReadAll(payload)
	if err != nil {
		return fmt.Errorf("Failed to read input, error: %s", err)
	}
	input := gojsonschema.NewStringLoader(string(data))

	// Validate against input schema
	result, err := gojsonschema.Validate(
		gojsonschema.NewStringLoader(schema), input,
	)
	if err != nil {
		return fmt.Errorf("Validation failed, error: %s", err)
	}

	// Print all validation errors
	for _, e := range result.Errors() {
		fmt.Fprintf(os.Stderr, " - %s\n", e.Description())
	}

	if !result.Valid() {
		return errors.New("Input is invalid")
	}
	return nil
}

func execute(
	baseURL string, entry *definitions.Entry, args, query map[string]string,
	payload io.Reader, output io.Writer,
) error {
	var input []byte
	// Read all input
	if entry.Input != "" {
		data, err := ioutil.ReadAll(payload)
		if err != nil {
			return fmt.Errorf("Failed to read input, error: %s", err)
		}
		input = data
	}

	// Parameterize the route
	route := entry.Route
	for k, v := range args {
		val := strings.Replace(url.QueryEscape(v), "+", "%20", -1)
		route = strings.Replace(route, "<"+k+">", val, 1)
	}

	// Create query options
	qs := make(url.Values)
	for k, v := range query {
		qs.Add(k, v)
	}
	q := qs.Encode()
	if q != "" {
		q = "?" + q
	}

	// Construct parameters
	method := strings.ToUpper(entry.Method)
	url := baseURL + route + q

	// Try to make the request up to 5 times using go-got
	// Allow unlimited responses.
	g := got.New()
	g.Retries = 5
	g.MaxSize = 0

	req := g.NewRequest(method, url, input)

	// If there is a body, we set a content-type
	if len(input) != 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	// Sign request if credentials are available
	if config.Credentials != nil {
		var h hash.Hash
		// Create payload hash if there is any
		if len(input) != 0 {
			h = client.PayloadHash("application/json")
			h.Write(input)
		}
		err := config.Credentials.SignGotRequest(req, h)
		if err != nil {
			return fmt.Errorf("Failed to sign request, error: %s", err)
		}
	}

	res, err := req.Send()
	if err != nil {
		return fmt.Errorf("Request failed: %s", err)
	}

	// Print the request to whatever output
	_, err = output.Write(res.Body)
	if err != nil {
		return fmt.Errorf("Failed to print response: %s", err)
	}

	// Exit
	return nil
}
