pushd protos
git pull origin master
popd
protoc  --go_out=. protos/**.proto
