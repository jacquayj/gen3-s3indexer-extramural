FROM golang:1.13.1 as build-deps

ENV GOPATH=/

RUN mkdir -p /src/github.com/jacquayj/gen3-s3indexer-extramural
WORKDIR /src/github.com/jacquayj/gen3-s3indexer-extramural
ADD . .

RUN go get -u github.com/aws/aws-sdk-go/aws
RUN go get -u github.com/jacquayj/indexs3client
RUN go install -tags netgo -ldflags '-extldflags "-static"' github.com/jacquayj/gen3-s3indexer-extramural
RUN go install -tags netgo -ldflags '-extldflags "-static"' github.com/jacquayj/indexs3client

# Store only the resulting binary in the final image
# Resulting in significantly smaller docker image size
FROM busybox
COPY --from=build-deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-deps /bin/gen3-s3indexer-extramural /gen3-s3indexer-extramural
COPY --from=build-deps /bin/indexs3client /indexs3client
COPY --from=build-deps /src/github.com/jacquayj/gen3-s3indexer-extramural/manifest.txt /manifest.txt

ENTRYPOINT /gen3-s3indexer-extramural
