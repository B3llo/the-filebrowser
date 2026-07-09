## Multistage build: First stage fetches dependencies
FROM alpine:3.23 AS fetcher

ARG TARGETARCH
ARG TARGETVARIANT

# install and copy ca-certificates, mailcap, and tini-static; download JSON.sh
RUN apk update && \
    apk --no-cache add ca-certificates mailcap tini-static && \
    wget -O /JSON.sh https://raw.githubusercontent.com/dominictarr/JSON.sh/0d5e5c77365f63809bf6e77ef44a1f34b0e05840/JSON.sh

# download a static ffmpeg build (musl/busybox has no package manager to install it from)
# and verify it against the upstream-published checksum before extracting.
RUN set -eu; \
    case "$TARGETARCH-$TARGETVARIANT" in \
        "amd64-") FFMPEG_ARCH="amd64" ;; \
        "arm64-") FFMPEG_ARCH="arm64" ;; \
        "arm-v7") FFMPEG_ARCH="armhf" ;; \
        "arm-v6") FFMPEG_ARCH="armel" ;; \
        *) echo "no static ffmpeg build for $TARGETARCH-$TARGETVARIANT, skipping" && FFMPEG_ARCH="" ;; \
    esac; \
    mkdir -p /usr/local/bin; \
    if [ -n "$FFMPEG_ARCH" ]; then \
        BASE_URL="https://johnvansickle.com/ffmpeg/releases"; \
        ARCHIVE="ffmpeg-release-$FFMPEG_ARCH-static.tar.xz"; \
        mkdir -p /ffmpeg-dl && cd /ffmpeg-dl; \
        wget -O "$ARCHIVE" "$BASE_URL/$ARCHIVE"; \
        wget -O "$ARCHIVE.md5" "$BASE_URL/$ARCHIVE.md5"; \
        md5sum -c "$ARCHIVE.md5"; \
        mkdir -p /ffmpeg-extract && tar -xJf "$ARCHIVE" -C /ffmpeg-extract; \
        find /ffmpeg-extract -type f -name ffmpeg -exec cp {} /usr/local/bin/ffmpeg \; ; \
        chmod +x /usr/local/bin/ffmpeg; \
    else \
        touch /usr/local/bin/ffmpeg; \
    fi

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
