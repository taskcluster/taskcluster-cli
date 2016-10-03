package s3

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/taskcluster/taskcluster-cli/config"
	"github.com/taskcluster/taskcluster-cli/extpoints"
	"github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/auth"
)

func init() {
	extpoints.Register("s3", S3{})
}

type S3 struct{}

func (S3) ConfigOptions() map[string]extpoints.ConfigOption {
	return nil
}

func (S3) Summary() string {
	return "Uploads/downloads file to/from AWS S3."
}

func usage() string {
	return `Upload/Download file to/from AWS S3.
Usage:
  taskcluster S3 put <file> <bucket> <prefix>
  taskcluster S3 get <key> <bucket> <prefix>
  taskcluster S3 help <subcommand>
`
}

func (S3) Usage() string {
	return usage()
}

func (S3) Execute(context extpoints.Context) bool {
	argv := context.Arguments

	// Check for command
	command := argv["S3"].(string)
	provider := extpoints.CommandProviders()[command]
	if provider == nil {
		panic(fmt.Sprintf("Unknown command: %s", command))
	}

	// Print help for subcommands
	if argv["help"] == true {
		subcommand, ok := argv["<subcommand>"].(string)
		if !ok {
			panic(fmt.Sprintf("Unknown subcommand: %s", subcommand))
		}
		if subcommand == "put" {
			fmt.Println("Usage:\ntaskcluster S3 put <file> <bucket> <prefix>")
		} else if subcommand == "get" {
			fmt.Println("Usage:\ntaskcluster S3 get <key> <bucket> <prefix>")
		} else {
			fmt.Printf("Invalid subcommand.\n")
		}
		return true
	}

	// Parse for bucket and prefix
	bucket, ok := argv["<bucket>"].(string)
	if !ok {
		fmt.Println("Invalid bucket format.")
		return false
	}
	prefix, ok := argv["<prefix>"].(string)
	if !ok {
		fmt.Println("Invalid prefix format.")
		return false
	}

	// Set level, filename and key
	var level, filename, key string
	if argv["put"] == true {
		level = "read-write"
		filename, ok = argv["<file>"].(string)
		if !ok {
			fmt.Println("Invalid file format.")
			return false
		}
	} else if argv["get"] == true {
		level = "read-write" // can change this later
		key, ok = argv["<key>"].(string)
		if !ok {
			fmt.Println("Invalid key format.")
			return false
		}
	}

	// Load configuration
	config, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load configuration file, error: ", err)
		return false
	}

	// Set credentials for auth
	myAuth := auth.New(
		&tcclient.Credentials{
			ClientID:    config["config"]["clientId"].(string),
			AccessToken: config["config"]["accessToken"].(string),
			Certificate: config["config"]["certificate"].(string),
		},
	)

	// Get credentials for AWS S3
	resp, err := myAuth.AwsS3Credentials(level, bucket, prefix)
	if err != nil {
		fmt.Println("Failed to load AWS S3 credentials: ", err)
		return false
	}

	// Set Credentials for AWS S3
	aws_access_key_id := resp.Credentials.AccessKeyID
	aws_secret_access_key := resp.Credentials.SecretAccessKey
	aws_session_token := resp.Credentials.SessionToken

	creds := credentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, aws_session_token)
	_, err = creds.Get()
	if err != nil {
		fmt.Printf("Invalid credentials: %s", err)
		return false
	}

	// Get bucket location
	aws_config := aws.NewConfig().WithCredentials(creds).WithRegion("us-west-2")
	sess := session.New(aws_config)
	svc := awsS3.New(sess)
	params := &awsS3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	}
	res, err := svc.GetBucketLocation(params)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	region := res.LocationConstraint

	// Create a session which contains the configurations for the SDK.
	// Use the session to create the service clients to make API calls to AWS.
	aws_config = aws.NewConfig().WithCredentials(creds).WithRegion(*region)
	sess = session.New(aws_config)
	svc = awsS3.New(sess)

	if argv["put"] == true {
		upload := Put(svc, filename, bucket, prefix)
		return upload
	} else if argv["get"] == true {
		download := Get(svc, key, bucket, prefix)
		return download
	} else {
		panic(fmt.Sprint("Invalid subcommand."))
	}
}

func Put(svc *awsS3.S3, filename string, bucket string, prefix string) bool {
	// Open the filename and return the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Failed to open file", filename, err)
		return false
	}
	defer file.Close()

	// Generate key from filename
	key := filepath.Base(filename)

	fmt.Printf("Uploading %s to s3://%s/%s...\n", key, bucket, prefix)

	// Upload the file
	params := &awsS3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(prefix + key),
		Body:   file,
	}
	_, err = svc.PutObject(params)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	fmt.Printf("Successfully uploaded %s to s3://%s/%s\n", key, bucket, prefix)
	return true
}

func Get(svc *awsS3.S3, key string, bucket string, prefix string) bool {
	// Get current path
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error occured: ", err)
		return false
	}

	// Create filepath
	filename := filepath.Join(pwd, key)
	if err := os.MkdirAll(filepath.Dir(filename), 0775); err != nil {
		fmt.Println("Unable to create directory: ", err)
		return false
	}

	// Setup the local file
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Printf("Downloading from s3://%s/%s to %s...\n", bucket, prefix+key, filename)

	// Download the file
	params := &awsS3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(prefix + key),
	}

	resp, err := svc.GetObject(params)
	if err != nil {
		fmt.Printf("Failed to download data to %s from s3://%s/%s\n", filename, bucket, prefix+key)
		return false
	}

	// Write data to the file
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading content", err)
	}
	file.Write(data)

	fmt.Printf("Successfully downloaded to %s from s3://%s/%s. (%d bytes)\n", filename, bucket, prefix+key, *resp.ContentLength)
	return true
}
