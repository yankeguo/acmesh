FROM ghcr.io/guoyk93/minit:1.9.1 AS minit

FROM alpine:3.17

RUN mkdir -p /opt/bin
ENV PATH "/opt/bin:${PATH}"

COPY --from=minit /minit /opt/bin/minit
ENTRYPOINT ["/opt/bin/minit"]

RUN apk add --no-cache curl tzdata openssl bash coreutils socat curl ca-certificates

RUN mkdir -p /acmesh.origin && \
    curl -sSL -o /acmesh.tar.gz 'https://github.com/acmesh-official/acme.sh/tarball/d4befeb5360e278fbc6be775c0ec60de464ed9fb' && \
    tar -xvf acmesh.tar.gz --strip-components 1 -C /acmesh.origin && \
    rm -f acmesh.tar.gz

RUN curl -sSL -o kubectl.tar.gz 'https://dl.k8s.io/v1.24.10/kubernetes-client-linux-amd64.tar.gz' && \
    tar -xvf kubectl.tar.gz --strip-components 3 -C /opt/bin && \
    rm -f kubectl.tar.gz

ADD scripts /opt/bin
ADD minit.d /etc/minit.d