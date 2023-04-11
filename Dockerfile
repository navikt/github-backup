FROM cgr.dev/chainguard/go:1.20 as build

WORKDIR /go/github-backup
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 make
RUN CGO_ENABLED=0 make test

FROM cgr.dev/chainguard/static
COPY --from=build /go/github-backup/bin /
CMD ["/github-backup"]