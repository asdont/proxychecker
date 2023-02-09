FROM golang:alpine AS builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod vendor -a -installsuffix cgo -o proxychecker cmd/app/main.go

FROM alpine:latest
WORKDIR /proxychecker
COPY --from=builder ./build/proxychecker .
COPY --from=builder ./build/configs/ /proxychecker/configs/
COPY --from=builder ./build/data/ /proxychecker/data/
COPY --from=builder ./build/docs/ /proxychecker/docs/
COPY --from=builder ./build/web/ /proxychecker/web/
CMD ["./proxychecker"]