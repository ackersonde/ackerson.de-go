FROM alpine:latest
EXPOSE 8080

RUN apk --no-cache add curl mailcap
ADD homepage /app/
WORKDIR /app

ENTRYPOINT ["/app/homepage"]
