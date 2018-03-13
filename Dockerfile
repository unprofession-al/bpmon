FROM alpine
RUN apk add --update \
    bash \
    curl
COPY bin/bpmon /bpmon
COPY hacking/docker/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["./entrypoint.sh"]
EXPOSE 8910
