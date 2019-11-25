# gen3-s3indexer-extramural

Indexes S3 data, even if the data existed before S3 events were configured for ssjdispatcher.

## Usage

```
{
    "jobDefinitionName": "gen3-indexer",
    "jobDefinitionArn": "arn:aws:batch:us-east-1:098381893833:job-definition/gen3-indexer:17",
    "revision": 17,
    "status": "ACTIVE",
    "type": "container",
    "parameters": {},
    "containerProperties": {
        "image": "index.docker.io/jacquayj/gen3-s3indexer-extramural:1.0.1",
        "vcpus": 4,
        "memory": 4000,
        "command": [
            "/bin/sh",
            "-c",
            "/gen3-s3indexer-extramural $AWS_BATCH_JOB_ARRAY_INDEX 50"
        ],
        "volumes": [],
        "environment": [
            {
                "name": "AWS_REGION",
                "value": "us-east-1"
            },
            {
                "name": "AWS_ACCESS_KEY_ID",
                "value": "-redacted-"
            },
            {
                "name": "NUM_WORKERS",
                "value": "10"
            },
            {
                "name": "AWS_SECRET_ACCESS_KEY",
                "value": "-redacted-"
            },
            {
                "name": "AWS_BUCKET",
                "value": "-redacted-"
            },
            {
                "name": "INDEXD_PASS",
                "value": "-redacted-"
            },
            {
                "name": "INDEXD_URL",
                "value": "-redacted-"
            },
            {
                "name": "INDEXD_UPLOADER",
                "value": "john@bioteam.net"
            },
            {
                "name": "INDEXD_USER",
                "value": "gdcapi"
            },
            {
                "name": "JOB_QUEUE_SIZE",
                "value": "1000"
            }
        ],
        "mountPoints": [],
        "ulimits": [],
        "resourceRequirements": []
    },
    "timeout": {
        "attemptDurationSeconds": 86400
    }
}
```
