FROM golang:1.27-alpine@sha256:d530cf60723abf1ffae0cd16bb546b379ec0550d2f3815e5b16ed69dbc00fba4 AS builder
RUN apk add make
ARG BUILD_VERSION
ENV BUILD_VERSION=${BUILD_VERSION}
ADD --chown=1001:0 . /tezosconnect
WORKDIR /tezosconnect
RUN mkdir /.cache \
    && chgrp -R 0 /.cache \
    && chmod -R g+rwX /.cache
USER 1001
RUN make

# Copy the migrations from FFTM down into our local migrations directory
RUN DB_MIGRATIONS_DIR=$(go list -f '{{.Dir}}' github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi | sed 's|pkg/ffcapi|db|') \
 && cp -R $DB_MIGRATIONS_DIR db

FROM debian:bookworm-slim@sha256:7ff6ba31527b9e336162fa061746c366c4b2032f4d2cff2078bea43f78efe3f8
WORKDIR /tezosconnect
RUN chgrp -R 0 /tezosconnect \
    && chmod -R g+rwX /tezosconnect
RUN apt update -y \
 && apt install -y curl jq \
 && rm -rf /var/lib/apt/lists/* \
 && curl -sL "https://github.com/golang-migrate/migrate/releases/download/$(curl -sL https://api.github.com/repos/golang-migrate/migrate/releases/latest | jq -r '.name')/migrate.linux-amd64.tar.gz" | tar xz \
 && chmod +x ./migrate \
 && mv ./migrate /usr/bin/migrate
COPY --from=builder --chown=1001:0 /tezosconnect/firefly-tezosconnect /usr/bin/tezosconnect
COPY --from=builder --chown=1001:0 /tezosconnect/db/ /tezosconnect/db/
USER 1001

ENTRYPOINT [ "/usr/bin/tezosconnect" ]
