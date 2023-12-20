FROM docker.io/library/alpine:latest
ENV RUNNING_IN_DOCKER=true
ENTRYPOINT ["/bin/bash"]
CMD ["/app/pvpc_exporter.sh"]
RUN apk add --quiet --no-cache bash curl jq