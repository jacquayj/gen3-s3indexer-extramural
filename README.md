# gen3-s3indexer-extramural

Indexes S3 data for Gen3's `indexd` microservice, fast. You're going to need a bigger RDS instance.

## Todo

* Move secrets from environment vars to AWS secret store

## Usage

First you need to generate a manifest containing the following example info used for job submissions:
```javascript
{
        "jobs": [
                {
                        "start_key": null,
                        "end_key": "dg.XXXX/2525cfe8-d233-4d0c-9601-0b69d222b2a5/clinical.json"
                },
                {
                        "start_key": "dg.XXXX/2525cfe8-d233-4d0c-9601-0b69d222b2a5/clinical.json",
                        "end_key": "dg.XXXX/8a097d37-e4c0-49a2-b433-728521a8cd2a/output.tsv"
                },
                {
                        "start_key": "dg.XXXX/8a097d37-e4c0-49a2-b433-728521a8cd2a/output.tsv",
                        "end_key": null
                }
        ],
        "opts": {
                "regexs": null,
                "prefix": "dg.XXXX",
                "batch_size": 3
        }
}
```

Download the tool used to generate this manifest file.
```sh
$ docker pull jacquayj/gen3-s3indexer-manifest
$ docker run jacquayj/gen3-s3indexer-manifest -h
Usage:
  manifest [OPTIONS]

Application Options:
  -r, --regex=      Object keys must match this or be skipped, multiple
                    expressions can be specified
  -p, --prefix=     Limits the response to keys that begin with the specified
                    prefix
  -s, --batch-size= Batch cluster size (default: 10)

Help Options:
  -h, --help        Show this help message

```

Clone this repo, the manifest you generate needs to be inside the `gen3-s3indexer-extramural` directory.
```
$ git clone https://github.com/jacquayj/gen3-s3indexer-extramural.git
$ cd gen3-s3indexer-extramural
```

Generate the `manifest.json` file: 

1. Save ENV file `.env` containing your AWS S3 configuration, see `.env.example` as reference:
```
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_REGION=us-east-1
AWS_BUCKET=
```

2. Pass in the desired `--batch-size`, and any prefixes (`--prefix`) or regex filters (`--regex`). Also specify the `.env` file you created.

```
$ docker run --env-file=.env jacquayj/gen3-s3indexer-manifest \
  --batch-size=3 \
  --prefix="dg.XXXX" > manifest.json
```

Then build the job container, including the `manifest.json` you generated in previous steps (should exist in same directory).
```
$ docker build -t username/my-batch-container .
$ docker push username/my-batch-container
```

## AWS Batch Usage

```
{
    "jobDefinitionName": "gen3-indexer",
    "jobDefinitionArn": "arn:aws:batch:us-east-1:098381893833:job-definition/gen3-indexer:17",
    "revision": 17,
    "status": "ACTIVE",
    "type": "container",
    "parameters": {},
    "containerProperties": {
        "image": "index.docker.io/username/my-batch-container:latest,
        "vcpus": 4,
        "memory": 4000,
        "command": [
            "/bin/sh",
            "-c",
            "/gen3-s3indexer-extramural $AWS_BATCH_JOB_ARRAY_INDEX"
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
