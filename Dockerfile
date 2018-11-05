# Base build image
FROM golang:1.11-alpine AS build_base

RUN apk add bash ca-certificates git gcc g++ libc-dev
WORKDIR /go/src/github.com/unprofession-al/bpmon

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

# App build image
FROM build_base AS app_builder

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go install -a -tags netgo -ldflags '-w -extldflags "-static"' ./...

# App image
FROM scratch
COPY --from=app_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=app_builder /go/bin/bpmon /bpmon
ENTRYPOINT ["./bpmon"]
