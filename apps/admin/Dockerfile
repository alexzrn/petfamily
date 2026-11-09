FROM golang:1.23-alpine AS build
WORKDIR /src
COPY apps/admin/go.mod apps/admin/go.sum ./
RUN go mod download
COPY apps/admin/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/admin .

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/admin /admin
EXPOSE 8081
ENTRYPOINT ["/admin"]
