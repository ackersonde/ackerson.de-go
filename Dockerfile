FROM alpine:latest
RUN apk --no-cache add curl mailcap
ADD homepage /app/
WORKDIR /app
ENTRYPOINT ["/app/homepage"]
