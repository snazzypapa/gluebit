# GlueBit: a qbittorrent/gluetun companion
GlueBit is a lightweight command-line app to keep your qbittorrent instance listening for connections on the right port.

It is useful when routing qbittorrent network traffic through a gluetun docker container.

## Install
### Option 1: go install
If you have go installed on your machine, run:
```
go install github.com/snazzypapa/gluebit@latest
```
### Option 2: Build with docker and run on the host (Linux only)
If you don't have go installed but you have docker installed, you can quickly build the application in a docker container and copy the binary to the host machine.

```
# clone the repo and change directory
git clone github.com/snazzypapa/gluebit.git && cd gluebit

# build the binary
sh dockerbuild.sh

# you should now have a binary file called gluebit in the folder

# change ownership to root 
sudo chown root:root ./gluebit

# move the binary to /usr/local/bin
sudo mv ./gluebit /usr/local/bin/gluebit
```
You may want to prune the build cache to free some disk space:
```
docker builder prune
```

### Option 3: To run with docker

Build the image:
```
# clone the repo and change directory
git clone github.com/snazzypapa/gluebit.git && cd gluebit

docker build -t gluebit .
```

## Usage
GlueBit collects the port forwarded by the VPN server from gluetun, either from the gluetun control api or from a file, and then assigns the listen port in qbittorrent via the webui API. To do so, GlueBit needs to know how to communicate with both services.

Arguments can be passed via command-line arguments or environment variables.

If no qbittorrent username or password is provided, GlueBit will try to login without password authorization.

```
Usage: gluebit [--qbituser QBITUSER] [--qbitpass QBITPASS] [--qbithost QBITHOST] [--qbitport QBITPORT] [--gluetunhost GLUETUNHOST] [--gluetunport GLUETUNPORT] [--gluetunportfile GLUETUNPORTFILE] [--interval INTERVAL]

Options:
  --qbituser QBITUSER    qbittorrent username [env: QBITUSER]
  --qbitpass QBITPASS    qbittorrent password [env: QBITPASS]
  --qbithost QBITHOST    host to reach qbittorrent on. If this is run on the same docker network as gluetun, this can be set to the container name [default: localhost, env: QBITHOST]
  --qbitport QBITPORT    port to reach qbittorrent on [default: 8080, env: QBITPORT]
  --gluetunhost GLUETUNHOST    host to reach gluetun on. If this is run on the same docker network as gluetun, this can be set to the container name [default: localhost, env: GLUETUNHOST]
  --gluetunport GLUETUNPORT    port to reach gluetun on [default: 8000, env: GLUETUNPORT]
  --gluetunportfile GLUETUNPORTFILE    path to gluetun port file [env: GLUETUNPORTFILE]
  --interval INTERVAL    Update interval in seconds [default: 60, env: INTERVAL]
  --help, -h             display this help and exit
```

### Run in docker
If you run GlueBit on the same docker network as gluetun, and qbittorrent is using your gluetun container's network, docker will resolve hosts by their container names. For instance, running on the network called 'saltbox':
```
docker run \
    --rm \
    --network saltbox \
    gluebit --qbithost gluetun --qbitport 8080 --gluetunhost gluetun --gluetunport 8000
```
Or with environment variables:
```
docker run \
    --rm \
    --name gluebit \
    --network saltbox \
    -e QBITUSER=admin \
    -e QBITPASS=adminadmin \
    -e QBITHOST=gluetun \
    -e GLUETUNHOST=gluetun \
    -e GLUETUNPORT=8000 \
    gluebit
```


