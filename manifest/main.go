package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jessevdk/go-flags"
)

const MANIFEST_FILE = "/manifest.txt"

var (
	AWS_ACCESS_KEY_ID     = os.Getenv("AWS_ACCESS_KEY_ID")
	AWS_SECRET_ACCESS_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")
	AWS_REGION            = os.Getenv("AWS_REGION")
	AWS_BUCKET            = os.Getenv("AWS_BUCKET")
)

var opts struct {
	Regexs    []string `short:"r" description:"Object keys must match this or be skipped"`
	Prefix    *string  `short:"p" description:"Limits the response to keys that begin with the specified prefix"`
	BatchSize int      `short:"s" description:"Batch cluster size" default:"10"`
}

type ParsedRegexes []*regexp.Regexp

func main() {

	_, err := flags.Parse(&opts)
	if err != nil {
		panic(err)
	}

	regexes := make(ParsedRegexes, len(opts.Regexs))
	for i, rStr := range opts.Regexs {
		regexes[i] = regexp.MustCompile(rStr)
	}

	sess := session.Must(session.NewSession())
	s3Svc := s3.New(sess, &aws.Config{
		Region:      aws.String(AWS_REGION),
		Credentials: credentials.NewStaticCredentials(AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, ""),
	})

	mf, err := os.Create(MANIFEST_FILE)
	if err != nil {
		panic(err)
	}

	bucket := AWS_BUCKET
	s3Svc.ListObjectsV2Pages(
		&s3.ListObjectsV2Input{Bucket: &bucket, Prefix: opts.Prefix},
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, obj := range page.Contents {
				objKey := *obj.Key
				if len(regexes) > 0 {
					for _, exp := range regexes {
						if exp.Match([]byte(objKey)) {
							fmt.Fprintln(mf, objKey)
							break
						}
					}
				} else {
					fmt.Fprintln(mf, objKey)
				}
			}
			return true
		},
	)

	if err := mf.Close(); err != nil {
		panic(err)
	}

	resp := Jobs{}

	// Could make this faster by batching all calls to single getKeysAtLines
	for i := 0; i < opts.BatchSize; i++ {
		resp.BatchRuns = append(resp.BatchRuns, calculateStartEndKeys(i, opts.BatchSize))
	}

	manifestJSON, err := json.MarshalIndent(resp, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(manifestJSON))
}
