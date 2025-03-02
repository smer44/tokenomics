#!/bin/bash

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Icons
INFO_ICON="💡"
SUCCESS_ICON="✅"
WARNING_ICON="⚠️ "
ERROR_ICON="❌"

# Hardcoded GitHub repository path
IMAGE_NAME="ghcr.io/smer44/tokenomics"

# Default port from Dockerfile is 8080
PORT=8080

# Default branch is the current git branch, fallback to main if not in a git repo
BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")

# Default to attached mode
DETACH=false

# Container name
CONTAINER_NAME="tokenomics"

# Cleanup function
cleanup() {
    echo -e "\n${WARNING_ICON} ${YELLOW}Received interrupt signal. Cleaning up...${NC}"
    if [ "$DETACH" = true ]; then
        echo -e "${INFO_ICON} ${BLUE}Stopping and removing container...${NC}"
        docker stop "$CONTAINER_NAME" >/dev/null 2>&1
        docker rm "$CONTAINER_NAME" >/dev/null 2>&1
    fi
    exit 0
}

# Register cleanup function for SIGINT and SIGTERM
trap cleanup SIGINT SIGTERM

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
            echo -e "${ERROR_ICON} ${RED}Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# Set default tag if not specified - use latest-<branch>
TAG=${TAG:-latest-$BRANCH}

# Clean branch name (replace invalid characters with dash)
TAG=$(echo "$TAG" | sed 's/[^a-zA-Z0-9]/-/g')

# Check if container already exists and remove it
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${INFO_ICON} ${BLUE}Found existing container. Removing it...${NC}"
    docker stop "$CONTAINER_NAME" >/dev/null 2>&1
    docker rm "$CONTAINER_NAME" >/dev/null 2>&1
fi

# Pull the latest version of the image
echo -e "${INFO_ICON} ${BLUE}Pulling image ${IMAGE_NAME}:${TAG}...${NC}"
docker pull "${IMAGE_NAME}:${TAG}"

# Prepare docker run command
DOCKER_CMD="docker run"
if [ "$DETACH" = true ]; then
    DOCKER_CMD="$DOCKER_CMD -d"
fi

# Print access information before starting the container
echo -e "\n${SUCCESS_ICON} ${GREEN}Starting container...${NC}"
echo -e "${INFO_ICON} ${BLUE}Access the application:${NC}"
echo -e "  📚 Swagger UI      : ${GREEN}http://localhost:${PORT}/docs${NC}"
echo -e "\n${INFO_ICON} ${BLUE}To view logs, run:${NC} ${GREEN}docker logs $CONTAINER_NAME${NC}"
echo -e "${INFO_ICON} ${YELLOW}Press Ctrl+C to stop and remove the container${NC}"

# Run the container
$DOCKER_CMD \
    --name "$CONTAINER_NAME" \
    -p "${PORT}:8080" \
    --restart unless-stopped \
    "${IMAGE_NAME}:${TAG}"

# Exit immediately in detached mode
if [ "$DETACH" = true ]; then
    exit 0
fi 