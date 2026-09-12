FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/denisurl .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/denisurl /denisurl

EXPOSE 8000
USER nonroot:nonroot
ENTRYPOINT ["/denisurl"]
