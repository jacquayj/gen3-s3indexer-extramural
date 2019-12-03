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
	"github.com/jacquayj/gen3-s3indexer-extramural/common"
	"github.com/jessevdk/go-flags"
)

const MANIFEST_FILE = "/manifest.txt"

var (
	AWS_ACCESS_KEY_ID     = os.Getenv("AWS_ACCESS_KEY_ID")
	AWS_SECRET_ACCESS_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")
	AWS_REGION            = os.Getenv("AWS_REGION")
	AWS_BUCKET            = os.Getenv("AWS_BUCKET")
)

var opts common.ManifestOpts

type ParsedRegexes []*regexp.Regexp

func main() {

	_, err := flags.Parse(&opts)
	if err != nil {
		if !flags.WroteHelp(err) {
			panic(err)
		}
		return
	}

	resp := common.Jobs{Opts: opts}

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
	listErr := s3Svc.ListObjectsV2Pages(
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
				resp.ObjCount++
			}
			return true
		},
	)
	if listErr != nil {
		panic(err)
	}

	if err := mf.Close(); err != nil {
		panic(err)
	}

	// Only calculate the lines to fetch from manifest file
	for i := 0; i < opts.BatchSize; i++ {
		resp.RawBatchRuns = append(resp.RawBatchRuns, calculateStartEndKeys(opts.BatchSize, i))
	}

	// Fetch the lines in one manifest loop
	resolveBatchRuns(&resp)

	manifestJSON, err := json.MarshalIndent(resp, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(manifestJSON))
}
