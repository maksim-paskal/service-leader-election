FROM alpine:latest
ARG TARGETARCH

COPY ./service-leader-election-${TARGETARCH} /app/service-leader-election

WORKDIR /app

RUN apk upgrade \
&& addgroup -g 31101 -S app \
&& adduser -u 31101 -D -S -G app app

USER 31101

ENTRYPOINT [ "/app/service-leader-election" ]