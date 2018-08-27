FROM golang:alpine as builder
MAINTAINER Daniel Menet

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates

COPY . /go/src/github.com/unprofession-al/bpmon

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		git \
		gcc \
		libc-dev \
		libgcc \
		make \
	&& go get -u github.com/mjibson/esc \
	&& cd /go/src/github.com/unprofession-al/bpmon \
	&& make \
	&& mv bpmon /usr/bin/bpmon \
	&& apk del .build-deps \
	&& rm -rf /go \
	&& echo "Build complete."

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/bin/bpmon /bpmon
ENTRYPOINT ["./bpmon"]
