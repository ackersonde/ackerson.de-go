FROM phusion/baseimage
ENV HOME /root
# Use baseimage-docker's init process.
CMD ["/sbin/my_init"]

# install Go
RUN apt-get install -y golang git
RUN mkdir /root/gocode
ENV GOPATH /root/gocode

# install ackerson.de
RUN git clone https://github.com/danackerson/ackerson.de-go.git /root/gocode/src/github.com/danackerson/ackerson.de-go/
WORKDIR /root/gocode/src/github.com/danackerson/ackerson.de-go
RUN go get ./...
RUN go build server.go
EXPOSE 3001

# execute ackerson.de
ENTRYPOINT ["/root/gocode/src/github.com/danackerson/ackerson.de-go/server"]
