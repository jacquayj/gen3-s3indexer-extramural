module github.com/jacquayj/gen3-s3indexer-extramural

go 1.13

replace github.com/jacquayj/gen3-s3indexer-extramural/common => ./common

require (
	github.com/aws/aws-sdk-go v1.25.43
	github.com/jacquayj/indexs3client v0.0.0-20191130141751-68bc59445cc1
	github.com/jessevdk/go-flags v1.4.1-0.20181221193153-c0795c8afcf4
)
