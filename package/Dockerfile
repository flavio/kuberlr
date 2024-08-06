# If desired this can be set to specific tag by passing an arg
ARG BCI_VERSION=latest
FROM registry.suse.com/bci/bci-busybox:${BCI_VERSION} AS final

# Image that provides cross compilation tooling.
FROM --platform=$BUILDPLATFORM rancher/mirrored-tonistiigi-xx:1.3.0 AS xx

FROM --platform=$BUILDPLATFORM registry.suse.com/bci/golang:1.22 AS build

ARG PROJECT_PATH
ARG VERSION
ARG BUILD_DATE
ARG TAG

COPY --from=xx / /

WORKDIR /app
COPY . .

ENV CGO_ENABLED=0
RUN xx-go --wrap && mkdir -p /run/lock
RUN xx-go mod download && \
    xx-go build -ldflags "-X=${PROJECT_PATH}/pkg/kuberlr.Version=${VERSION} \
        -X=${PROJECT_PATH}/pkg/kuberlr.BuildDate=${BUILD_DATE:-$(date +%Y%m%d)} \
        -X=${PROJECT_PATH}/pkg/kuberlr.Tag=${TAG}" \
        -o ./bin/kuberlr ./cmd/kuberlr
RUN xx-verify --static ./bin/kuberlr

FROM registry.suse.com/bci/bci-base:${BCI_VERSION} AS zypper

RUN mkdir /chroot
COPY --from=final / /chroot/

# The final image does not contain zypper, --installroot is used to
# install all artefacts within a dir (/chroot) that can then be copied
# over to a scratch image.
RUN zypper --non-interactive refresh && \
    zypper --installroot /chroot -n rm busybox-vi busybox-links && \
    zypper --installroot /chroot -n in bash-completion && \
    zypper --installroot /chroot clean -a && \
    rm -rf /chroot/var/cache/zypp/* /chroot/var/log/zypp/* /chroot/etc/zypp/

COPY --from=build /app/bin/kuberlr /chroot/bin/kuberlr
RUN cd /chroot/bin && ln -s ./kuberlr ./kubectl

RUN useradd -u 1000 -U -m kuberlr \
    && cp /etc/passwd /chroot/etc/passwd \
    && cp /etc/group /chroot/etc/group \
    && cp -r /home/kuberlr /chroot/home/kuberlr \
    && echo '. /etc/profile.d/bash_completion.sh' >> /chroot/home/kuberlr/.bashrc \
    && echo 'alias k="kubectl"' >> /chroot/home/kuberlr/.bashrc \
    && echo 'alias ks="kubectl -n kube-system"' >> /chroot/home/kuberlr/.bashrc \
    && echo 'source <(kubectl completion bash)' >> /chroot/home/kuberlr/.bashrc \
    && echo 'LANG=en_US.UTF-8' >> /chroot/home/kuberlr/.bashrc \
    && echo 'PS1="> "' >> /chroot/home/kuberlr/.bashrc \
    && mkdir /chroot/home/kuberlr/.kube \
    && mkdir /chroot/home/kuberlr/.kuberlr \
    && touch /chroot/home/kuberlr/.kuberlr/kuberlr.conf \
    && echo "AllowDownload = true" >> /chroot/home/kuberlr/.kuberlr/kuberlr.conf \
    && echo "Timeout = 12" >> /chroot/home/kuberlr/.kuberlr/kuberlr.conf \
    && chown -R 1000:1000 /chroot/home/kuberlr

FROM scratch
COPY --from=zypper /chroot /

USER kuberlr

WORKDIR /home/kuberlr
ENTRYPOINT ["/bin/kuberlr"]
CMD ["help"]