FROM ghcr.io/guoyk93/acicn/alpine:3.16

RUN apk add --no-cache openssl bash coreutils socat curl ca-certificates

RUN mkdir -p /acmesh.origin && \
    curl -sSL -o /acmesh.tar.gz 'https://github.com/acmesh-official/acme.sh/archive/refs/tags/v3.0.5.tar.gz' && \
    tar -xvf acmesh.tar.gz --strip-components 1 -C /acmesh.origin && \
    rm -f acmesh.tar.gz

ADD scripts /opt/bin
ADD minit.d /etc/minit.d