FROM golang:1.19.0-alpine3.16 as build
ARG OS
ARG ARCH
COPY . /build/
WORKDIR /build
RUN RUN go mod download && go build -o godaddyddns

FROM alpine:3.16
ARG VERSION
ARG user=godaddyddns
ARG group=godaddyddns
ARG uid=1000
ARG gid=1000
USER root
WORKDIR /app
COPY --from=build /build/godaddyddns /app/godaddyddns
COPY container-entrypoint.sh /app/container-entrypoint.sh
RUN apk update && apk --no-cache add bash && addgroup -g ${gid} ${group} && adduser -h /app -u ${uid} -G ${group} -s /bin/bash -D ${user}
RUN chown godaddyddns:godaddyddns /app/godaddyddns && chmod +x /app/godaddyddns && \
    chown godaddyddns:godaddyddns /app/container-entrypoint.sh && chmod +x /app/container-entrypoint.sh
USER godaddyddns
ENTRYPOINT [ "/app/container-entrypoint.sh"]