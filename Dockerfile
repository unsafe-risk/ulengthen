FROM golang:1.19-alpine as build
WORKDIR /build
COPY . .

RUN go build -ldflags "-s -w" -o /app.exe .

FROM scratch
COPY --from=build /app.exe /app.exe
ENTRYPOINT ["/app.exe"]
