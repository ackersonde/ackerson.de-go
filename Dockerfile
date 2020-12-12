FROM multiarch/alpine:armv7-latest-stable
RUN apk --no-cache add curl mailcap
ADD homepage /app/
WORKDIR /app
ENTRYPOINT ["/app/homepage"]
