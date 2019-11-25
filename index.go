package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
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

	AWS_BATCH_JOB_ARRAY_INDEX = os.Getenv("AWS_BATCH_JOB_ARRAY_INDEX")
	AWS_BATCH_JOB_ARRAY_SIZE  = os.Getenv("AWS_BATCH_JOB_ARRAY_SIZE")
)

var indexS3ClientConfig = struct {
	URL                   string `json:"url"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	ExtramuralBucket      bool   `json:"extramural_bucket"`
	ExtramuralUploader    string `json:"extramural_uploader"`
	ExtramuralInitialMode bool   `json:"extramural_initial_mode"`
}{
	os.Getenv("INDEXD_URL"),
	os.Getenv("INDEXD_USER"),
	os.Getenv("INDEXD_PASS"),
	true,
	os.Getenv("INDEXD_UPLOADER"),
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

type BatchRun struct {
	StartKey *string
	EndKey   *string
}

func calculateStartEndKeys() BatchRun {
	numTotalObjs, err := getManifestNumLines("/manifest.txt")
	if err != nil {
		log.Fatal(err)
	}
	objsPerNode := int(math.Ceil(float64(numTotalObjs) / float64(batchSize)))

	startLine := batchIndex * objsPerNode
	endLine := (batchIndex + 1) * objsPerNode

	if batchIndex == 0 {
		return BatchRun{
			nil,
			getKeyAtLine("/manifest.txt", endLine),
		}
	}

	if (batchIndex + 1) == batchSize {
		return BatchRun{
			getKeyAtLine("/manifest.txt", startLine),
			nil,
		}
	}

	key1, key2 := getKeyAtLine2("/manifest.txt", startLine, endLine)
	return BatchRun{
		key1,
		key2,
	}
}

func main() {

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
