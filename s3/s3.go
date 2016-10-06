package s3

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
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
  taskcluster S3 get <bucket> <prefix> [<target>]
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
			fmt.Println("Usage:\ntaskcluster S3 get <bucket> <prefix> [<target>]")
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

	// Set level, filename and target
	var level, filename, target string
	if argv["put"] == true {
		level = "read-write"
		filename, ok = argv["<file>"].(string)
		if !ok {
			fmt.Println("Invalid file format.")
			return false
		}
	} else if argv["get"] == true {
		level = "read-only"
		target, ok = argv["<target>"].(string)
		if !ok {
			target = ""
		}
	}

	// Set credentials
	authCreds := tcclient.Credentials(*context.Credentials)
	myAuth := auth.New(&authCreds)

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
	region, err := findBucketRegion(creds, bucket)
	if err != nil {
		fmt.Println("Failed to find bucket region", err)
		return false
	}

	// Create a session which contains the configurations for the SDK.
	// Use the session to create the service clients to make API calls to AWS.
	aws_config := aws.NewConfig().WithCredentials(creds).WithRegion(*region)
	sess := session.New(aws_config)
	svc := awsS3.New(sess)

	if argv["put"] == true {
		return put(svc, filename, bucket, prefix)
	} else if argv["get"] == true {
		return get(svc, bucket, prefix, target)
	} else {
		panic(fmt.Sprint("Invalid subcommand."))
	}
}

func findBucketRegion(creds *credentials.Credentials, bucket string) (region *string, err error) {
	aws_config := aws.NewConfig().WithCredentials(creds).WithRegion("us-east-1")
	sess := session.New(aws_config)
	svc := awsS3.New(sess)
	params := &awsS3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	}
	res, err := svc.GetBucketLocation(params)
	if err != nil {
		return nil, err
	}
	region = res.LocationConstraint
	return
}

func put(svc *awsS3.S3, filename string, bucket string, prefix string) bool {
	// Open the filename
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Failed to open file", filename, err)
		return false
	}
	defer file.Close()

	// If prefix ends in a slash, add last element of path
	key := prefix
	endsWithSlash := strings.HasSuffix(prefix, "/")
	if endsWithSlash {
		key = prefix + filepath.Base(filename)
	}

	fmt.Printf("Uploading to s3://%s/%s...\n", bucket, prefix)

	// Upload the file
	params := &awsS3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	}
	_, err = svc.PutObject(params)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	fmt.Printf("Successfully uploaded %s to s3://%s/%s\n", filepath.Base(filename), bucket, prefix)
	return true
}

func get(svc *awsS3.S3, bucket string, prefix string, target string) bool {
	// Get current path
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error occured: ", err)
		return false
	}

	// If target is not present, use current directory
	var filename string
	if target != "" {
		filename = target
	} else {
		filename = filepath.Join(pwd, filepath.Base(prefix))
	}

	// Setup the local file
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Printf("Downloading from s3://%s/%s...\n", bucket, prefix)

	// Download the file
	params := &awsS3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(prefix),
	}

	resp, err := svc.GetObject(params)
	if err != nil {
		fmt.Printf("Failed to download data from s3://%s/%s\n", bucket, prefix)
		return false
	}

	// Write data to the file, retry if error occurs
	maxTries := 5
	for maxTries > 0 {
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fmt.Println("Error reading content", err)
			maxTries--
			if maxTries != 0 {
				fmt.Println("Trying again...")
			}
			continue
		} else {
			break
		}
	}

	fmt.Printf("Successfully downloaded to %s from s3://%s/%s. (%d bytes)\n", filename, bucket, prefix, *resp.ContentLength)
	return true
}
