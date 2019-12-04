module github.com/jacquayj/gen3-s3indexer-extramural

go 1.13

replace github.com/jacquayj/gen3-s3indexer-extramural/common => ./common

require (
	github.com/aws/aws-sdk-go v1.25.47
	github.com/jacquayj/indexs3client v0.0.0-20191204212648-cfc16e948a83
	github.com/jessevdk/go-flags v1.4.1-0.20181221193153-c0795c8afcf4
)
