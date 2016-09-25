package s3

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
	awsS3Manager "github.com/aws/aws-sdk-go/service/s3/s3manager"
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
  taskcluster S3 upload <file> <region> <level> <bucket> <prefix>
  taskcluster S3 download <key> <region> <level> <bucket> <prefix>
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

		if subcommand == "upload" {
			fmt.Println(`Usage:
  taskcluster S3 upload <file> <region> <level> <bucket> <prefix>`)
		} else if subcommand == "download" {
			fmt.Println(`Usage:
  taskcluster S3 download <key> <region> <level> <bucket> <prefix>`)
		}
		return true
	}

	// Parse for bucket, prefix, level and region
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
	level, ok := argv["<level>"].(string)
	if !ok {
		fmt.Println("Invalid level format.")
		return false
	}
	region, ok := argv["<region>"].(string)
	if !ok {
		fmt.Println("Invalid region format.")
		return false
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

	// Get credential for AWS S3
	resp, err := myAuth.AwsS3Credentials(level, bucket, prefix)
	if err != nil {
		fmt.Println("Failed to load AWS S3 credentials: ", err)
		return false
	}

	// Set Credentials for AWS S3
	aws_access_key_id := resp.Credentials.AccessKeyID
	aws_secret_access_key := resp.Credentials.SecretAccessKey
	aws_session_token := resp.Credentials.SessionToken
	//aws_expires := &resp.Expires

	creds := credentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, aws_session_token)
	_, err = creds.Get()
	if err != nil {
		fmt.Printf("Invalid credentials: %s", err)
		return false
	}

	// Create a session which contains the configurations for the SDK.
	// Use the session to create the service clients to make API calls to AWS.
	aws_config := aws.NewConfig().WithRegion(region).WithCredentials(creds)
	sess := session.New(aws_config)

	// For upload
	if argv["upload"] == true {

		// Parse for filename
		filename, ok := argv["<file>"].(string)
		if !ok {
			fmt.Println("Invalid file format.")
			return false
		}

		// Create service client
		svc := awsS3Manager.NewUploader(sess)

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
		params := &awsS3Manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   file,
			//Expires: aws_expires,
		}
		_, err = svc.Upload(params)
		if err != nil {
			if multierr, ok := err.(awsS3Manager.MultiUploadFailure); ok {
				// Process error and its associated uploadID for Multipart Upload failure
				fmt.Println("Multipart Upload error: ", multierr.Code(), multierr.Message(), multierr.UploadID())
				return false
			} else {
				// Process error generically
				fmt.Printf("Failed to upload %s to s3://%s/%s, %s\n", key, bucket, prefix, err.Error())
				return false
			}
		}

		fmt.Printf("Successfully uploaded %s to s3://%s/%s\n", key, bucket, prefix)

	} else if argv["download"] == true {

		// Parse for key
		key, ok := argv["<key>"].(string)
		if !ok {
			fmt.Println("Invalid key format.")
			return false
		}

		// Create service client
		svc := awsS3Manager.NewDownloader(sess)

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

		fmt.Printf("Downloading from s3://%s/%s to %s...\n", bucket, key, filename)

		// Download the file
		params := &awsS3.GetObjectInput{Bucket: &bucket, Key: &key}
		_, err = svc.Download(file, params)
		if err != nil {
			fmt.Printf("Failed to download data to %s from s3://%s/%s\n", filename, bucket, key)
			return false
		}

		fmt.Printf("Successfully downloaded to %s from s3://%s/%s\n", filename, bucket, key)
	} else {
		panic(fmt.Sprintf("Unknown command: %s", command))
	}
	return true
}
