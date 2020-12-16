module github.com/jacquayj/gen3-s3indexer-extramural

go 1.13

replace github.com/jacquayj/gen3-s3indexer-extramural/common => ./common

require (
	github.com/aws/aws-sdk-go v1.36.9
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/jacquayj/indexs3client v0.0.0-20201216171332-9a20e5fead53
	github.com/jessevdk/go-flags v1.4.1-0.20181221193153-c0795c8afcf4
)
