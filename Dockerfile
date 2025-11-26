# syntax=docker/dockerfile:1.3-labs

FROM debian:bookworm-slim AS base
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update -y && \
    apt-get install -y --no-install-recommends \
    curl tini ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# We use bookworm since the icu dependency ver. between the base and golang images is the same 
FROM golang:1.25-bookworm AS build-from-source
ENV DEBIAN_FRONTEND=noninteractive
ARG DOLTGRES_VERSION="latest"

RUN mkdir -p /tmp/doltgresql/
COPY . /tmp/doltgresql/
WORKDIR /tmp/doltgresql/

RUN if [ "$DOLTGRES_VERSION" = "source" ]; then \
    go mod download; \
    ./scripts/build_binaries.sh "linux-amd64"; \
    mv out/doltgresql-*/bin/doltgres /usr/local/bin; \
    fi

FROM base AS download-binary
ARG DOLTGRES_VERSION="latest"

RUN if [ "$DOLTGRES_VERSION" = "latest" ]; then \
    DOLTGRES_VERSION=$(curl -s "https://api.github.com/repos/dolthub/doltgresql/releases/latest" \
      | grep '"tag_name"' \
      | cut -d'"' -f4 \
      | sed 's/^v//'); \
    echo "fetching https://github.com/dolthub/doltgresql/releases/download/v${DOLTGRES_VERSION}/install.sh"; \
    curl -L "https://github.com/dolthub/doltgresql/releases/download/v${DOLTGRES_VERSION}/install.sh" | bash; \
    fi

RUN if [ "$DOLTGRES_VERSION" != "latest" ] && [ "$DOLTGRES_VERSION" != "source" ]; then \
    echo "fetching https://github.com/dolthub/doltgresql/releases/download/v${DOLTGRES_VERSION}/install.sh"; \
    curl -L "https://github.com/dolthub/doltgresql/releases/download/v${DOLTGRES_VERSION}/install.sh" | bash; \
    fi   

FROM base AS runtime

# Only one binary is possible due to DOLT_VERSION, so we optionally copy from either stage
COPY --from=download-binary /usr/local/bin/dolt* /usr/local/bin/
COPY --from=build-from-source /usr/local/bin/dolt* /usr/local/bin/

RUN /usr/local/bin/doltgres --version

RUN mkdir /docker-entrypoint-initdb.d && \
    mkdir -p /var/lib/doltgres && \
    chmod 755 /var/lib/doltgres

COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

VOLUME /var/lib/doltgres
# TODO: are all these ports on doltgres?
EXPOSE 5432 33060 7007
WORKDIR /var/lib/doltgres
ENTRYPOINT ["tini", "--", "docker-entrypoint.sh"]
