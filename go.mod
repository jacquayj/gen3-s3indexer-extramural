module github.com/jacquayj/gen3-s3indexer-extramural

go 1.13

replace github.com/jacquayj/gen3-s3indexer-extramural/common => ./common

require (
	github.com/aws/aws-sdk-go v1.34.32
	github.com/jacquayj/indexs3client v0.0.0-20200925195612-8be590d8e71c
	github.com/jessevdk/go-flags v1.4.1-0.20181221193153-c0795c8afcf4
)
