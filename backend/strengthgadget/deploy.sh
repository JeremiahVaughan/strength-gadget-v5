#!/bin/bash
echo "$DOCKER_TOKEN" | docker login -u "$DOCKER_USER" --password-stdin
docker buildx create --use
docker buildx build --platform linux/arm64 --push -t "$DOCKER_USER/strengthgadget:$CIRCLE_WORKFLOW_ID" .