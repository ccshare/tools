#!/bin/sh

app_name="fud"
app_version="1.0.1"

release_dir=./release
rm -rf $release_dir
mkdir -p $release_dir

cd  $(dirname $0)

gofmt -w ./

# "linux" "darwin" "freebsd" "windows"
for goos in "linux" "darwin" "windows"
do
    if [ "$goos" = "windows" ]
    then
      obj_name=$app_name.exe
      GOOS=$goos GOARCH=amd64 go build
      zip $release_dir/$app_name-$goos-amd64-$app_version.zip $obj_name
    elif [ "$goos" = "linux" ]
    then
      obj_name=$app_name
      GOOS=$goos GOARCH=amd64 go build
      zip $release_dir/$app_name-$goos-amd64-$app_version.zip $obj_name
      GOOS=$goos GOARCH=arm go build
      zip $release_dir/$app_name-$goos-arm-$app_version.zip $obj_name
      GOOS=$goos GOARCH=arm64 go build
      zip $release_dir/$app_name-$goos-arm64-$app_version.zip $obj_name
    else
      obj_name=$app_name
      GOOS=$goos GOARCH=amd64 go build
      zip $release_dir/$app_name-$goos-amd64-$app_version.zip $obj_name
    fi

    rm -f $obj_name
done

cd $release_dir
sha1sum *.zip >> sha1sum.txt

