FROM multiarch/alpine:armv7-v3.10
RUN apk --no-cache add curl mailcap
ADD homepage /app/
WORKDIR /app
ENTRYPOINT ["/app/homepage"]
