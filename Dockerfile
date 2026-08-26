FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/buildfarm ./cmd/buildfarm
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/buildfarm /buildfarm
EXPOSE 8080
ENTRYPOINT ["/buildfarm"]
