FROM phusion/baseimage
ENV HOME /root
# Use baseimage-docker's init process.
CMD ["/sbin/my_init"]
RUN apt-get update
RUN apt-get upgrade -y

# install Go
RUN apt-get install -y golang git
RUN mkdir /root/gocode
ENV GOPATH /root/gocode

# install ackerson.de
RUN go get github.com/codegangsta/negroni
RUN go get github.com/goincremental/negroni-session
RUN go get gopkg.in/mgo.v2
RUN git clone https://github.com/danackerson/ackerson.de-go.git /root/gocode/src/github.com/danackerson/ackerson.de-go/
WORKDIR /root/gocode/src/github.com/danackerson/ackerson.de-go
RUN go build server.go
EXPOSE 3001

# execute ackerson.de
ENTRYPOINT ["/root/gocode/src/github.com/danackerson/ackerson.de-go/server"]
