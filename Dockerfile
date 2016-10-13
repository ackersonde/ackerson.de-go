FROM alpine:latest

# install Go
RUN mkdir -p /root/gocode
ENV GOPATH /root/gocode
RUN apk add -U git go

# install ackerson.de
RUN git clone https://github.com/danackerson/ackerson.de-go.git $GOPATH/src/github.com/danackerson/ackerson.de-go/
WORKDIR $GOPATH/src/github.com/danackerson/ackerson.de-go

RUN go get ./...
RUN go test

RUN go build server.go
RUN mv server /root/
RUN mkdir /root/certs

RUN apk del git go && \
  rm -rf $GOPATH/pkg && \
  rm -rf $GOPATH/bin && \
  rm -rf $GOPATH/src/gopkg.in && \
  rm -rf $GOPATH/src/github.com/clbanning && \
  rm -rf $GOPATH/src/github.com/urfave && \
  rm -rf $GOPATH/src/github.com/goincremental && \
  rm -rf $GOPATH/src/github.com/gorilla && \
  rm -rf $GOPATH/src/github.com/unrolled && \
  rm -rf /var/cache/apk/*

# execute ackerson.de
ENTRYPOINT ["/root/server"]
