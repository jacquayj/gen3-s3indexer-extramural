module github.com/jacquayj/gen3-s3indexer-extramural

go 1.13

replace github.com/jacquayj/gen3-s3indexer-extramural/common => ./common

replace github.com/jacquayj/gen3-s3indexer-extramural/manifest => ./manifest

require (
	github.com/aws/aws-sdk-go v1.25.43
	github.com/jacquayj/gen3-s3indexer-extramural/common v0.0.0-00010101000000-000000000000
	github.com/jacquayj/indexs3client v0.0.0-20191130141751-68bc59445cc1
)
