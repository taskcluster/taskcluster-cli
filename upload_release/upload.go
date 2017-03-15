package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	tcclient "github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/auth"
)

func getSTSCredentials() (id, secret, token string, err error) {
	// TODO: make sure this works with proxy
	creds := tcclient.Credentials{}
	client := auth.New(&creds)
	resp, err := client.AwsS3Credentials("read-write", "downloads-taskcluster-net", "taskcluster-cli", "")
	if err != nil {
		return "", "", "", err
	}
	return resp.Credentials.AccessKeyID,
		resp.Credentials.SecretAccessKey,
		resp.Credentials.SessionToken,
		nil
}

func getS3() (*s3.S3, error) {
	id, secret, token, err := getSTSCredentials()
	if err != nil {
		return nil, err
	}

	creds := credentials.NewStaticCredentials("id", "secret", "token")
	sess := session.Must(session.NewSession(aws.NewConfig().
		WithRegion("us-west-1").
		WithCredentials(creds)))

	return s3.New(sess), nil
}

func main() {
	svc, err := getS3()
	if err != nil {
		exitErrorf("Unable to get service object, %v", err)
	}

	result, err := svc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list buckets, %v", err)
	}

	fmt.Println("Buckets:")
	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
