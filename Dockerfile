FROM alpine:latest
RUN apk --no-cache add curl mailcap
ADD server /app/
WORKDIR /app
ENTRYPOINT ["/app/server"]
