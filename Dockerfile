ARG GO_VERSION="1.25"
FROM golang:${GO_VERSION} AS builder
WORKDIR /src
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o github-backup .

FROM scratch
COPY --from=builder /src/github-backup /github-backup
CMD ["/github-backup"]