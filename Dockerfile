FROM golang:1.18 as build

WORKDIR /go/github-backup
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 make
RUN CGO_ENABLED=0 make test

FROM gcr.io/distroless/static-debian11
COPY --from=build /go/github-backup/bin /
CMD ["/github-backup"]