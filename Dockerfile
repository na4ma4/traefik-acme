FROM scratch
ARG DOCKER_PLATFORM=linux/amd64
LABEL org.opencontainers.image.source https://github.com/na4ma4/traefik-acme
COPY artifacts/build/release/${DOCKER_PLATFORM}/traefik-acme /
ENTRYPOINT [ "/traefik-acme" ]
