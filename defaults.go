package main

import (
	"log"
	"strconv"
	"strings"
)

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

}
