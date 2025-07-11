# --- Build Stage ---
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

WORKDIR /app

# Update apk and install git for go mod if needed
RUN apk update && apk upgrade && apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Set target platform variables for cross-compilation
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app/cmd
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/argocd-namespace-generator .

# --- Final Stage ---
FROM --platform=$TARGETPLATFORM scratch

COPY --from=builder /out/argocd-namespace-generator /argocd-namespace-generator

ENTRYPOINT ["/argocd-namespace-generator"]