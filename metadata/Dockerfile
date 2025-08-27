FROM ubuntu:22.04

WORKDIR /app
COPY ./cmd/main ./
COPY ./configs/defaults.yaml ./
RUN chmod +x ./main
EXPOSE 8081
CMD ["/app/main"]