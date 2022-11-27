# go-wake-on-lan-server

### Deploying:
`docker-compose build && docker-compose up -d`

### Updating:
```bash
docker-compose build \
  && docker-compose down \
  && docker-compose up -d
```
## Build:
arm64: `docker buildx build --platform linux/arm64 -t go-wake-on-lan-server:arm64 --output="type=docker,push=false,name=$REPOSITORY:$TAG,dest=go-wake-on-lan-server_arm64.tar" .`
amd64: `docker buildx build --platform linux/amd64 -t go-wake-on-lan-server:amd64 --output="type=docker,push=false,name=$REPOSITORY:$TAG,dest=go-wake-on-lan-server_amd64.tar" .`
