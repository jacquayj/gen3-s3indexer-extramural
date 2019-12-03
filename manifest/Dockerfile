FROM golang:1.13.4 as build-deps

RUN mkdir -p /build
WORKDIR /build
ADD . .

RUN go build -o manifest -tags netgo -ldflags '-extldflags "-static"' .

# Store only the resulting binary in the final image
# Resulting in significantly smaller docker image size
FROM busybox
COPY --from=build-deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-deps /build/manifest /manifest

ENTRYPOINT [ "/manifest" ] 