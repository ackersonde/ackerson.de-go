FROM alpine:latest
RUN apk --no-cache add curl
ADD . /app/
WORKDIR /app
ENTRYPOINT ["/app/server"]
