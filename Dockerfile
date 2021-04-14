FROM alpine:latest
EXPOSE 8080

RUN apk --no-cache add curl mailcap

WORKDIR /app
ADD homepage /app/

ENTRYPOINT ["/app/homepage"]
