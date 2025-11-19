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
ARG DOLTGRES_VERSION

RUN echo $(ls -l)
RUN echo $(pwd)
COPY . /tmp/doltgresql/
WORKDIR /tmp/doltgresql/

# Check for source to avoid unnecessary installation of build dependencies
RUN if [ "$DOLGRES_VERSION" = "source" ]; then \
        cd /tmp/doltgresql/ || { echo "Make sure the `doltgresql/` directory exists in your workspace to build from source."; exit 1; }; \
        apt-get update -y && \
        apt-get install -y libicu-dev && \
        rm -rf /var/lib/apt/lists/*; \
    fi

# Separate layers to avoid redundant downloads
RUN if [ "$DOLTGRES_VERSION" = "source" ]; then \
        go mod download; \
        ./scripts/build_binaries.sh; \
    fi

FROM base AS download-binary
ARG DOLTGRES_VERSION
RUN if [ "$DOLTGRES_VERSION" = "latest" ]; then \
        # Fetch latest version number from GitHub API
        DOLTGRES_VERSION=$(curl -s https://api.github.com/repos/dolthub/doltgresql/releases/latest \
            | grep '"tag_name"' \
            | cut -d'"' -f4 \
            | sed 's/^v//'); \
    fi && \
    if [ "$DOLTGRES_VERSION" != "source" ]; then \
        curl -L "https://github.com/dolthub/doltgresql/releases/download/v${DOLTGRES_VERSION}/install.sh" | bash; \
    fi


FROM base AS runtime
ARG DOLTGRES_VERSION

RUN apt-get update -y && apt-get install -y --no-install-recommends bzip2 gzip xz-utils zstd \
  && rm -rf /var/lib/apt/lists/*
# icu dependency for source builds
RUN if [ "$DOLTGRES_VERSION" = "source" ]; then \
        apt-get update -y && \
        apt-get install -y --no-install-recommends libicu-dev && \
        rm -rf /var/lib/apt/lists/*; \
    fi

# Only one binary is possible due to DOLTGRES_VERSION, so we optionally copy from either stage
COPY --from=download-binary /usr/local/bin/dolt* /usr/local/bin/
COPY --from=build-from-source /usr/local/bin/dolt* /usr/local/bin/

RUN /usr/local/bin/dolt version

RUN mkdir /docker-entrypoint-initdb.d && \
    mkdir -p /var/lib/doltgres && \
    chmod 755 /var/lib/doltgres

COPY docker*/docker-entrypoint*.sh /usr/local/bin/
COPY dolt*/docker*/docker-entrypoint*.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

VOLUME /var/lib/doltgres
EXPOSE 5432 33060 7007
WORKDIR /var/lib/doltgres
ENTRYPOINT ["tini", "--", "docker-entrypoint.sh"]
