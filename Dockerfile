FROM golang:1.20 AS build
ENV CGO_ENABLED 0
WORKDIR /go/src/app
ADD . .
RUN go build -o /acmesh-apply-secret ./cmd/acmesh-apply-secret
RUN go build -o /acmesh-upload-qcloud ./cmd/acmesh-upload-qcloud

FROM ghcr.io/guoyk93/minit:1.10.1 AS minit

FROM alpine:3.17

RUN mkdir -p /opt/bin
ENV PATH "/opt/bin:${PATH}"

COPY --from=minit /minit /opt/bin/minit
ENTRYPOINT ["/opt/bin/minit"]

RUN apk add --no-cache curl tzdata openssl bash coreutils socat curl ca-certificates

RUN mkdir -p /acmesh.origin && \
    curl -sSL -o /acmesh.tar.gz 'https://github.com/acmesh-official/acme.sh/tarball/b7caf7a0165d80dd1556b16057a06bb32025066d' && \
    tar -xvf acmesh.tar.gz --strip-components 1 -C /acmesh.origin && \
    rm -f acmesh.tar.gz

COPY --from=build /acmesh-apply-secret /opt/bin/acmesh-apply-secret
COPY --from=build /acmesh-upload-qcloud /opt/bin/acmesh-upload-qcloud

ADD scripts /opt/bin
ADD minit.d /etc/minit.d
