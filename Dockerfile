## Static ffmpeg binaries, pulled straight from mwader/static-ffmpeg (an
## actively maintained multi-arch image of hardened static PIE builds) via
## plain COPY --from below. This replaces wget-ing a tarball from
## johnvansickle.com's personal, non-CDN server, which is a known single
## point of build failure (intermittent errors, no SLA). Only amd64/arm64
## are published upstream; other targets keep the previous graceful
## fallback (an empty placeholder binary).
FROM mwader/static-ffmpeg:8.1.2-amd64 AS ffmpeg-amd64
FROM mwader/static-ffmpeg:8.1.2-arm64 AS ffmpeg-arm64

## Multistage build: First stage fetches dependencies
FROM alpine:3.23 AS fetcher

ARG TARGETARCH
ARG TARGETVARIANT

# install and copy ca-certificates, mailcap, and tini-static; download JSON.sh
RUN apk update && \
    apk --no-cache add ca-certificates mailcap tini-static && \
    wget -O /JSON.sh https://raw.githubusercontent.com/dominictarr/JSON.sh/0d5e5c77365f63809bf6e77ef44a1f34b0e05840/JSON.sh

COPY --from=ffmpeg-amd64 /ffmpeg /ffmpeg-amd64
COPY --from=ffmpeg-arm64 /ffmpeg /ffmpeg-arm64

# select the static ffmpeg binary matching the target architecture
RUN set -eu; \
    mkdir -p /usr/local/bin; \
    case "$TARGETARCH-$TARGETVARIANT" in \
        "amd64-") cp /ffmpeg-amd64 /usr/local/bin/ffmpeg ;; \
        "arm64-") cp /ffmpeg-arm64 /usr/local/bin/ffmpeg ;; \
        *) echo "no static ffmpeg build for $TARGETARCH-$TARGETVARIANT, skipping" && touch /usr/local/bin/ffmpeg ;; \
    esac; \
    rm -f /ffmpeg-amd64 /ffmpeg-arm64; \
    chmod +x /usr/local/bin/ffmpeg

## Second stage: Use lightweight BusyBox image for final runtime environment
FROM busybox:1.37.0-musl

# Define non-root user UID and GID
ENV UID=1000
ENV GID=1000
# Points at the static ffmpeg binary copied in below; set explicitly since
# busybox images don't reliably have /usr/local/bin on PATH.
ENV FB_FFMPEG_PATH=/usr/local/bin/ffmpeg

# Create user group and user
RUN addgroup -g $GID user && \
    adduser -D -u $UID -G user user

# Copy binary, scripts, and configurations into image with proper ownership
COPY --chown=user:user filebrowser /bin/filebrowser
COPY --chown=user:user docker/common/ /
COPY --chown=user:user docker/alpine/ /
COPY --chown=user:user --from=fetcher /sbin/tini-static /bin/tini
COPY --from=fetcher /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg
COPY --from=fetcher /JSON.sh /JSON.sh
COPY --from=fetcher /etc/ca-certificates.conf /etc/ca-certificates.conf
COPY --from=fetcher /etc/ca-certificates /etc/ca-certificates
COPY --from=fetcher /etc/mime.types /etc/mime.types
COPY --from=fetcher /etc/ssl /etc/ssl

# Create data directories, set ownership, and ensure healthcheck script is executable
RUN mkdir -p /config /database /srv && \
    chown -R user:user /config /database /srv \
    && chmod +x /healthcheck.sh

# Define healthcheck script
HEALTHCHECK --start-period=2s --interval=5s --timeout=3s CMD /healthcheck.sh

# Set the user, volumes and exposed ports
USER user

VOLUME /srv /config /database

EXPOSE 80

ENTRYPOINT [ "tini", "--", "/init.sh" ]
