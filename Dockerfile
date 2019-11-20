FROM alpine

WORKDIR /go/bin

COPY bin/storage-crawler /go/bin

CMD /go/bin/storage-crawler
