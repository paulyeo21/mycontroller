FROM --platform=linux/amd64 golang:1.18.1 AS builder
WORKDIR /workspace
COPY . ./
ENV GOPROXY=direct
RUN go mod download
RUN CGO_ENABLED=0 GO11MODULE=on go build -o mycontroller main.go
FROM public.ecr.aws/amazonlinux/amazonlinux:2
WORKDIR /
COPY --from=builder /workspace/mycontroller .
ENTRYPOINT ["/mycontroller"]
