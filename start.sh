#!/bin/bash

# Hardcoded GitHub repository path
IMAGE_NAME="ghcr.io/smer44/tokenomics"

# Default port from Dockerfile is 8080
PORT=8080

# Default branch is the current git branch, fallback to main if not in a git repo
BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")

# Default to attached mode
DETACH=false

# Help message
show_help() {
    echo "Usage: ./start.sh [OPTIONS]"
    echo "Run the tokenomics Docker container"
    echo
    echo "Options:"
    echo "  -p, --port PORT    Specify the port to expose (default: 8080)"
    echo "  -t, --tag TAG      Specify the Docker image tag (default: latest-<current-branch>)"
    echo "  -b, --branch NAME  Specify the branch to use (default: current git branch or main)"
    echo "  -d, --detach       Run container in background (detached mode)"
    echo "  -h, --help         Show this help message"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -b|--branch)
            BRANCH="$2"
            shift 2
            ;;
        -d|--detach)
            DETACH=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Set default tag if not specified - use latest-<branch>
TAG=${TAG:-latest-$BRANCH}

# Clean branch name (replace invalid characters with dash)
TAG=$(echo "$TAG" | sed 's/[^a-zA-Z0-9]/-/g')

# Pull the latest version of the image
echo "Pulling image ${IMAGE_NAME}:${TAG}..."
docker pull "${IMAGE_NAME}:${TAG}"

# Prepare docker run command
DOCKER_CMD="docker run"
if [ "$DETACH" = true ]; then
    DOCKER_CMD="$DOCKER_CMD -d"
fi

# Run the container
echo "Starting container on port ${PORT}..."
$DOCKER_CMD \
    --name tokenomics \
    -p "${PORT}:8080" \
    --restart unless-stopped \
    "${IMAGE_NAME}:${TAG}"

echo "Container started! The service is available at http://localhost:${PORT}"
echo "To view logs, run: docker logs tokenomics"
echo "To stop the container, run: docker stop tokenomics && docker rm tokenomics" 