package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	AWS_ACCESS_KEY_ID     = os.Getenv("AWS_ACCESS_KEY_ID")
	AWS_SECRET_ACCESS_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")
	AWS_REGION            = os.Getenv("AWS_REGION")
	AWS_BUCKET            = os.Getenv("AWS_BUCKET")

	NUM_WORKERS       = os.Getenv("NUM_WORKERS")
	JOB_QUEUE_SIZE    = os.Getenv("JOB_QUEUE_SIZE")
	INDEXS3CLIENT_BIN = os.Getenv("INDEXS3CLIENT_BIN")

	AWS_BATCH_JOB_ARRAY_INDEX = os.Args[1]
	AWS_BATCH_JOB_ARRAY_SIZE  = os.Args[2]

	INDEXD_URL      = os.Getenv("INDEXD_URL")
	INDEXD_USER     = os.Getenv("INDEXD_USER")
	INDEXD_PASS     = os.Getenv("INDEXD_PASS")
	INDEXD_UPLOADER = os.Getenv("INDEXD_UPLOADER")
)

var indexS3ClientConfig = struct {
	URL                   string `json:"url"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	ExtramuralBucket      bool   `json:"extramural_bucket"`
	ExtramuralUploader    string `json:"extramural_uploader"`
	ExtramuralInitialMode bool   `json:"extramural_initial_mode"`
}{
	INDEXD_URL,
	INDEXD_USER,
	INDEXD_PASS,
	true,
	INDEXD_UPLOADER,
	true,
}

func invokeIndexS3Client(env []string) {
	cmd := exec.Command(INDEXS3CLIENT_BIN)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Print(err)
	}
	cmd.Wait()
}

var wg sync.WaitGroup

func worker(id int, jobs <-chan func()) {
	for job := range jobs {
		job()
	}
}

func main() {
	fmt.Printf("Starting with os.Args: %v", os.Args)

	defaults()

	startEndKeys := calculateStartEndKeys()

	sess := session.Must(session.NewSession())
	svc := s3.New(sess, &aws.Config{
		Region:      aws.String(AWS_REGION),
		Credentials: credentials.NewStaticCredentials(AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, ""),
	})

	config, _ := json.Marshal(indexS3ClientConfig)
	indexSettings := []string{
		fmt.Sprintf("AWS_REGION=%s", AWS_REGION),
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", AWS_SECRET_ACCESS_KEY),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", AWS_SECRET_ACCESS_KEY),
		fmt.Sprintf("CONFIG_FILE=%s", string(config)),
	}

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
				settingsWithObj := append(indexSettings, fmt.Sprintf("INPUT_URL=%s", objURL))

				// Send job to workers
				wg.Add(1)
				jobs <- func() {
					invokeIndexS3Client(settingsWithObj)
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
