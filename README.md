##
A little Go program designed to run on an Onion Omega to control 12v automotive relays.

### Build & Deploy

`-ldflags="-s -w"` strips the debugging information

```
GOOS=linux GOARCH=mipsle go build -ldflags="-s -w" -o cruisercontrol lights.go
scp cruisercontrol root@192.168.3.1:/tmp/app
```