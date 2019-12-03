package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jacquayj/gen3-s3indexer-extramural/common"
	"github.com/jacquayj/indexs3client/handlers"
)

const (
	MANIFEST_FILE = "/manifest.json"
)

var (
	AWS_ACCESS_KEY_ID     = os.Getenv("AWS_ACCESS_KEY_ID")
	AWS_SECRET_ACCESS_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")
	AWS_REGION            = os.Getenv("AWS_REGION")
	AWS_BUCKET            = os.Getenv("AWS_BUCKET")

	NUM_WORKERS    = os.Getenv("NUM_WORKERS")
	JOB_QUEUE_SIZE = os.Getenv("JOB_QUEUE_SIZE")

	AWS_BATCH_JOB_ARRAY_INDEX = os.Args[1]

	INDEXD_URL      = os.Getenv("INDEXD_URL")
	INDEXD_USER     = os.Getenv("INDEXD_USER")
	INDEXD_PASS     = os.Getenv("INDEXD_PASS")
	INDEXD_UPLOADER = os.Getenv("INDEXD_UPLOADER")
	FAST_MODE       = os.Getenv("FAST_MODE")
)

var indexS3ClientConfig = handlers.IndexdInfo{
	URL:                   INDEXD_URL,
	Username:              INDEXD_USER,
	Password:              INDEXD_PASS,
	ExtramuralBucket:      true,
	ExtramuralUploader:    &INDEXD_UPLOADER,
	ExtramuralInitialMode: true,
	ExtramuralFastMode:    strings.TrimSpace(strings.ToLower(FAST_MODE)) == "true",
}

func invokeIndexS3Client(objURL string) {
	handlers.IndexS3ObjectEmbedded(objURL, &indexS3ClientConfig, &aws.Config{
		Region: aws.String(AWS_REGION),
		Credentials: credentials.NewStaticCredentials(
			AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, ""),
	})
}

var wg sync.WaitGroup

func worker(id int, jobs <-chan func()) {
	for job := range jobs {
		job()
	}
}

func main() {
	log.Printf("Starting with os.Args: %v", os.Args)

	defaults()

	// Parse manifest.json file to get startEndKeys
	manifestB, err := ioutil.ReadFile(MANIFEST_FILE)
	if err != nil {
		log.Fatal(err)
	}
	batchJobs := common.Jobs{}
	if err := json.Unmarshal(manifestB, &batchJobs); err != nil {
		log.Fatal(err)
	}

	startEndKeys := batchJobs.BatchRuns[batchIndex]

	sess := session.Must(session.NewSession())
	svc := s3.New(sess, &aws.Config{
		Region:      aws.String(AWS_REGION),
		Credentials: credentials.NewStaticCredentials(AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, ""),
	})

	// Setup worker pool
	jobs := make(chan func(), jobQueueSize)
	for w := 1; w <= numWorkers; w++ {
		go worker(w, jobs)
	}

	bucket := AWS_BUCKET
	svc.ListObjectsV2Pages(
		&s3.ListObjectsV2Input{Bucket: &bucket, StartAfter: startEndKeys.StartKey},
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, obj := range page.Contents {

				objURL := fmt.Sprintf("s3://%s/%s", AWS_BUCKET, *obj.Key)

				// Send job to workers
				wg.Add(1)
				jobs <- func() {
					invokeIndexS3Client(objURL)
					wg.Done()
				}

				// Stop processing
				if startEndKeys.EndKey != nil {
					if *obj.Key == *startEndKeys.EndKey {
						return false
					}
				}
			}
			return true
		},
	)
	close(jobs) // No more jobs to create

	// Wait until we've processed all objects
	wg.Wait()

}
