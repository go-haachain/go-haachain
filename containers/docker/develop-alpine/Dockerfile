FROM alpine:3.7

RUN \
  apk add --update go git make gcc musl-dev linux-headers ca-certificates && \
  git clone --depth 1 https://github.com/haachain/go-haachain && \
  (cd go-haaereum && make ghaa) && \
  cp go-haaereum/build/bin/ghaa /ghaa && \
  apk del go git make gcc musl-dev linux-headers && \
  rm -rf /go-haaereum && rm -rf /var/cache/apk/*

EXPOSE 8545
EXPOSE 30303

ENTRYPOINT ["/ghaa"]
