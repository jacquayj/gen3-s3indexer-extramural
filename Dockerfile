FROM golang:1.13.4 as build-deps


RUN mkdir -p /build
WORKDIR /build
ADD . .

RUN go build -o gen3-s3indexer-extramural -tags netgo -ldflags '-extldflags "-static"' .

# Store only the resulting binary in the final image
# Resulting in significantly smaller docker image size
FROM busybox
COPY --from=build-deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-deps /build/gen3-s3indexer-extramural /gen3-s3indexer-extramural
COPY --from=build-deps /build/manifest.json /manifest.json

CMD /gen3-s3indexer-extramural
