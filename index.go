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

const (
	NUM_WORKERS    = 5
	JOB_QUEUE_SIZE = 1000
)

var (
	ACCESS            = os.Getenv("AWS_ACCESS_KEY_ID")
	SECRET            = os.Getenv("AWS_SECRET_ACCESS_KEY")
	BUCKET            = os.Getenv("AWS_BUCKET")
	REGION            = os.Getenv("AWS_REGION")
	INDEXS3CLIENT_BIN = os.Getenv("INDEXS3CLIENT_BIN")
)

var indexS3ClientConfig = struct {
	URL                string `json:"url"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	ExtramuralBucket   bool   `json:"extramural_bucket"`
	ExtramuralUploader string `json:"extramural_uploader"`
}{
	os.Getenv("INDEXD_URL"),
	os.Getenv("INDEXD_USER"),
	os.Getenv("INDEXD_PASS"),
	true,
	os.Getenv("UPLOADER"),
}

func invokeIndexS3Client(objectURL string, env []string) {
	if INDEXS3CLIENT_BIN == "" {
		INDEXS3CLIENT_BIN = "./indexs3client"
	}

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

func worker(id int, jobs <-chan func()) {
	for job := range jobs {
		job()
	}
}

func main() {

	sess := session.Must(session.NewSession())
	svc := s3.New(sess, &aws.Config{
		Region:      aws.String(REGION),
		Credentials: credentials.NewStaticCredentials(ACCESS, SECRET, ""),
	})

	config, _ := json.Marshal(indexS3ClientConfig)
	indexSettings := []string{
		fmt.Sprintf("AWS_REGION=%s", REGION),
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", ACCESS),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", SECRET),
		fmt.Sprintf("CONFIG_FILE=%s", string(config)),
	}

	// Setup worker pool
	var wg sync.WaitGroup
	jobs := make(chan func(), JOB_QUEUE_SIZE)
	for w := 1; w <= NUM_WORKERS; w++ {
		go worker(w, jobs)
	}

	bucket := BUCKET
	svc.ListObjectsV2Pages(
		&s3.ListObjectsV2Input{Bucket: &bucket},
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, obj := range page.Contents {
				objURL := fmt.Sprintf("s3://%s/%s", BUCKET, *obj.Key)
				settingsWithObj := append(indexSettings, fmt.Sprintf("INPUT_URL=%s", objURL))

				// Send job to workers
				wg.Add(1)
				jobs <- func() {
					invokeIndexS3Client(objURL, settingsWithObj)
					wg.Done()
				}
			}
			return true
		},
	)
	close(jobs) // No more jobs to create

	// Drain the job queue
	wg.Wait()

}
