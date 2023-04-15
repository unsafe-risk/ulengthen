FROM golang:1.20-alpine as build
WORKDIR /build
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /app.exe .

FROM gcr.io/distroless/static-debian11
COPY --from=build /app.exe /
ENTRYPOINT ["/app.exe"]
