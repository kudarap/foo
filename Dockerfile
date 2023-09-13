FROM golang:1.21 as builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    # Injects build version, commit hash, and date on to binary.
    -ldflags="-X main.vTag=`cat VERSION` -X main.vCommit=`git rev-parse HEAD` -X main.vBuilt=`date -u +%s`" \
    -a -o app ./cmd/foosvc

FROM ubuntu:22.04
WORKDIR /usr/local/bin
RUN apt-get update && apt-get install -y build-essential bash ca-certificates
RUN update-ca-certificates
COPY --from=builder /src/app .

EXPOSE 80

CMD ["/usr/local/bin/app", "server"]
