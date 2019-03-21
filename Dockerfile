FROM alpine:latest
RUN apk --no-cache add curl mailcap
ADD . /app/
WORKDIR /app
ENTRYPOINT ["/app/server"]
