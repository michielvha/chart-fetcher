FROM alpine:latest

# Set build-time variable, imported from pipeline environment during `Build` step
ARG IMAGE_NAME

# Create a non-root user with a fixed UID and group ID
RUN addgroup -g 1000 ${IMAGE_NAME} && \
    adduser -D -u 1000 -G ${IMAGE_NAME} ${IMAGE_NAME}

# Copy the binary into the container and adjust permissions
COPY ${IMAGE_NAME} /usr/local/bin/${IMAGE_NAME}
RUN chmod +x /usr/local/bin/${IMAGE_NAME} && chown ${IMAGE_NAME}:${IMAGE_NAME} /usr/local/bin/${IMAGE_NAME}

# Switch to the non-root user
USER ${IMAGE_NAME}

# Set the default command to run the binary
ENTRYPOINT ["/usr/local/bin/${IMAGE_NAME}"]
