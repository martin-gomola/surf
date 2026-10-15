FROM golang:1.26.5-bookworm AS backend-source

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates \
      curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download

COPY VERSION COMPATIBILITY_VERSION ./
COPY protocol ./protocol
COPY backend ./backend

FROM backend-source AS backend-test
RUN cd backend && go test ./...

FROM backend-source AS builder

ARG SURF_VERSION
ARG SURF_COMPATIBILITY=1
ARG SURF_CLIENT_DEB_SHA256

RUN set -eux; \
    client_deb="/tmp/surf-client.deb"; \
    curl -fL "https://github.com/seg6/surf/releases/download/v${SURF_VERSION}/space.seg6.surf_${SURF_VERSION}-1_iphoneos-arm.deb" \
      -o "$client_deb"; \
    echo "${SURF_CLIENT_DEB_SHA256}  ${client_deb}" | sha256sum -c -; \
    cp "$client_deb" backend/internal/web/client/client.deb; \
    cd backend; \
    go run ./tools/clientbundlemeta \
      --deb "$client_deb" \
      --version "$SURF_VERSION" \
      --compatibility "$SURF_COMPATIBILITY" \
      --output internal/web/client/client.json; \
    CGO_ENABLED=0 go build -trimpath -tags=client_bundle \
      -ldflags="-s -w -X surf-backend/internal/config.AppVersion=${SURF_VERSION} -X surf-backend/internal/config.CompatibilityVersion=${SURF_COMPATIBILITY}" \
      -o /out/surf ./cmd/surf

FROM debian:bookworm-slim AS client-core-test

RUN apt-get update && apt-get install -y --no-install-recommends \
      build-essential \
      cmake \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY protocol ./protocol
COPY client/core ./client/core
RUN cmake -S client/core -B /tmp/client-core \
      -DCMAKE_BUILD_TYPE=Debug -DSURF_CORE_WARNINGS_AS_ERRORS=ON \
    && cmake --build /tmp/client-core \
    && ctest --test-dir /tmp/client-core --output-on-failure

FROM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates \
      chromium \
      curl \
      fonts-liberation \
      fonts-noto-color-emoji \
      fonts-noto-core \
      xvfb \
    && rm -rf /var/lib/apt/lists/*

ENV SURF_HOME=/data/surf \
    PROFILE=/data/profile \
    DOWNLOADS=/data/downloads \
    UPLOADS=/data/uploads \
    CHROME=/usr/bin/chromium \
    PORT=8080

COPY --from=builder /out/surf /usr/local/bin/surf
COPY packaging/docker/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

RUN chmod 0755 /usr/local/bin/surf /usr/local/bin/docker-entrypoint.sh \
    && mkdir -p /data/surf /data/profile /data/downloads /data/uploads

VOLUME ["/data/surf", "/data/profile", "/data/downloads", "/data/uploads"]
EXPOSE 8080 18081

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
