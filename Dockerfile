ARG GO_VERSION=1.23

FROM golang:${GO_VERSION}-alpine AS build

ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

WORKDIR /src
COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -trimpath \
	-ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
	-o /out/quuid ./cmd/quuid

FROM scratch

LABEL org.opencontainers.image.title="quuid"
LABEL org.opencontainers.image.description="Strict RFC 9562 UUIDs and larger identifiers for Go"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/MeViksry/quuid"

COPY --from=build /out/quuid /quuid

USER 65532:65532
ENTRYPOINT ["/quuid"]
