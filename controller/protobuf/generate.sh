


docker container run --name protoc --rm \
  -v "$(pwd)"/src:/go/src \
  -it protoc:0.2.1;
