FROM golang:1.25 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/api ./cmd/api

# distroless static 自带 CA 证书和 tzdata；scratch 没有，
# 一旦连 TLS 数据库（sslmode=require）或调用外部 HTTPS 会直接 x509 失败。
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/api /api

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/api"]
