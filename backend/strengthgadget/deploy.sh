#!/bin/bash
echo "$TF_VAR_docker_token" | docker login -u "$TF_VAR_docker_user" --password-stdin
docker buildx create --use
docker buildx build --platform linux/arm64 --push -t "$TF_VAR_docker_user/strengthgadget:$TF_VAR_circle_workflow_id" .