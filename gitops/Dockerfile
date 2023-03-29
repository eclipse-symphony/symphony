FROM mcr.microsoft.com/oss/go/microsoft/golang:1.20 AS builder

WORKDIR /app

COPY . .

RUN go build  -o bin/gitops ./cmd/gitops

FROM mcr.microsoft.com/oss/go/microsoft/golang:1.20 AS runner
WORKDIR /app

COPY --from=builder /app/bin/gitops ./gitops
EXPOSE 8080
ENTRYPOINT ["/app/gitops"]

