FROM golang:1.13.1 as build-deps

ENV GOPATH=/

RUN mkdir -p /src/github.com/jacquayj/gen3-s3indexer-extramural
WORKDIR /src/github.com/jacquayj/gen3-s3indexer-extramural
ADD . .

RUN go get -u github.com/aws/aws-sdk-go/aws

RUN go install -tags netgo -ldflags '-extldflags "-static"' github.com/jacquayj/gen3-s3indexer-extramural

# Store only the resulting binary in the final image
# Resulting in significantly smaller docker image size
FROM scratch
COPY --from=build-deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-deps /bin/gen3-s3indexer-extramural /gen3-s3indexer-extramural

CMD ["/gen3-s3indexer-extramural"]
