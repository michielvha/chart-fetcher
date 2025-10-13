FROM alpine:latest

# - Set build-time variables -
# imported from pipeline environment during `Build` step
ARG IMAGE_NAME
# automatically set by Docker buildx
ARG TARGETARCH
ARG TARGETOS

# Copy the ARG value to an ENV variable that will persist at runtime
ENV IMAGE_NAME=${IMAGE_NAME}

# Create a non-root user with a fixed UID and group ID
RUN addgroup -g 1000 ${IMAGE_NAME} && \
    adduser -D -u 1000 -G ${IMAGE_NAME} ${IMAGE_NAME}

# Copy the binary from the dist directory based on target architecture
# GoReleaser creates binaries in dist/${IMAGE_NAME}_${TARGETOS}_${TARGETARCH}_v1/ for amd64
# and dist/${IMAGE_NAME}_${TARGETOS}_${TARGETARCH}/ for arm64
COPY dist/${IMAGE_NAME}_${TARGETOS}_${TARGETARCH}*/${IMAGE_NAME} /usr/local/bin/${IMAGE_NAME}
RUN chmod +x /usr/local/bin/${IMAGE_NAME} && chown ${IMAGE_NAME}:${IMAGE_NAME} /usr/local/bin/${IMAGE_NAME}

# Switch to the non-root user & set working directory
USER ${IMAGE_NAME}
WORKDIR /home/${IMAGE_NAME}

# Use exec form with environment variable substitution
ENTRYPOINT ["/bin/sh", "-c", "/usr/local/bin/${IMAGE_NAME}"]
