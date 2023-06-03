FROM golang:1.20 AS build
ENV CGO_ENABLED 0
WORKDIR /go/src/app
ADD . .
RUN go build -o /acmesh-apply-secret ./cmd/acmesh-apply-secret

FROM ghcr.io/guoyk93/minit:1.10.1 AS minit

FROM alpine:3.17

RUN mkdir -p /opt/bin
ENV PATH "/opt/bin:${PATH}"

COPY --from=minit /minit /opt/bin/minit
ENTRYPOINT ["/opt/bin/minit"]

RUN apk add --no-cache curl tzdata openssl bash coreutils socat curl ca-certificates

RUN mkdir -p /acmesh.origin && \
    curl -sSL -o /acmesh.tar.gz 'https://github.com/acmesh-official/acme.sh/tarball/0d25f7612bf37a42d9c4fcb1abc493b2a5e495c3' && \
    tar -xvf acmesh.tar.gz --strip-components 1 -C /acmesh.origin && \
    rm -f acmesh.tar.gz

COPY --from=build /acmesh-apply-secret /opt/bin/acmesh-apply-secret

ADD scripts /opt/bin
ADD minit.d /etc/minit.d
