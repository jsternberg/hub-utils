FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.2.1 AS xx

FROM --platform=$BUILDPLATFORM golang:1.20-alpine AS golatest

FROM golatest AS gobuild-base
RUN apk add --no-cache clang lld musl-dev pkgconfig git
COPY --link --from=xx / /
WORKDIR /src
ENV CGO_ENABLED=0

FROM gobuild-base AS git-history
ARG TARGETPLATFORM
RUN --mount=target=. \
    --mount=target=/go/pkg/mod,type=cache \
    --mount=target=/root/.cache/go-build,type=cache \
    xx-go build -o /usr/bin/git-history ./cmd/git-history

FROM gobuild-base AS git-prune-branch
ARG TARGETPLATFORM
RUN --mount=target=. \
    --mount=target=/go/pkg/mod,type=cache \
    --mount=target=/root/.cache/go-build,type=cache \
    xx-go build -o /usr/bin/git-prune-branch ./cmd/git-prune-branch

FROM scratch AS binaries
COPY --link --from=git-history /usr/bin/git-history /
COPY --link --from=git-prune-branch /usr/bin/git-prune-branch /
COPY --link ./cmd/git-main /
COPY --link ./cmd/git-fixup /
