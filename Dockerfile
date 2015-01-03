FROM golang

ADD . /go/src/github.com/resourced/resourced

WORKDIR /go/src/github.com/resourced/resourced

RUN go get ./...

RUN go install github.com/resourced/resourced

CMD /go/bin/resourced

EXPOSE 55555