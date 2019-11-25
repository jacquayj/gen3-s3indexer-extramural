package main

import (
	"log"
	"os"
	"strconv"
	"strings"
)

var numWorkers, jobQueueSize, batchIndex, batchSize int

func defaults() {
	if INDEXS3CLIENT_BIN == "" {
		INDEXS3CLIENT_BIN = "./indexs3client"
	}
	if JOB_QUEUE_SIZE == "" {
		JOB_QUEUE_SIZE = "1000"
	}
	if NUM_WORKERS == "" {
		NUM_WORKERS = "10"
	}
	var err error
	numWorkers, err = strconv.Atoi(NUM_WORKERS)
	if err != nil {
		log.Fatal(err)
	}
	jobQueueSize, err = strconv.Atoi(JOB_QUEUE_SIZE)
	if err != nil {
		log.Fatal(err)
	}
	if jobQueueSize == 0 || numWorkers == 0 {
		log.Fatal("JOB_QUEUE_SIZE or NUM_WORKERS == 0, that won't work!")
	}
	required := []string{
		AWS_ACCESS_KEY_ID,
		AWS_SECRET_ACCESS_KEY,
		AWS_REGION,
		AWS_BUCKET,
	}
	for _, val := range required {
		if strings.TrimSpace(val) == "" {
			log.Fatal("AWS config is required but not set")
		}
	}

	requiredIndexd := []string{
		os.Getenv("INDEXD_URL"),
		os.Getenv("INDEXD_USER"),
		os.Getenv("INDEXD_PASS"),
		os.Getenv("INDEXD_UPLOADER"),
	}
	for _, val := range requiredIndexd {
		if strings.TrimSpace(val) == "" {
			log.Fatal("indexd config is required but not set")
		}
	}

	requiredBatch := []string{
		AWS_BATCH_JOB_ARRAY_INDEX,
		AWS_BATCH_JOB_ARRAY_SIZE,
	}
	for _, val := range requiredBatch {
		if strings.TrimSpace(val) == "" {
			log.Fatal("batch config is required but not set")
		}
	}
	batchIndex, err = strconv.Atoi(AWS_BATCH_JOB_ARRAY_INDEX)
	if err != nil {
		log.Fatal(err)
	}
	batchSize, err = strconv.Atoi(AWS_BATCH_JOB_ARRAY_SIZE)
	if err != nil {
		log.Fatal(err)
	}

}
