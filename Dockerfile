FROM multiarch/alpine:arm64-latest-stable
RUN apk --no-cache add curl mailcap
ADD homepage /app/
WORKDIR /app
ENTRYPOINT ["/app/homepage"]
