# Adapted from https://stackoverflow.com/a/53557155

# build the binary
docker build -t gluebit:build -f Dockerfile .
docker container create --name gluebit_build_temp gluebit:build

# copy the compiled binary from the container to the host cwd
docker container cp gluebit_build_temp:/gluebit ./

# remove generated resources
docker container rm gluebit_build_temp
docker image rm gluebit:build