FROM iron/base
RUN apk --no-cache add curl
RUN mkdir -p /root/certs
ADD . /app/
COPY server.pem /root/certs
COPY server.key /root/certs
WORKDIR /app
ENTRYPOINT ["/app/server"]
