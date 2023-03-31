#!/bin/bash

# $1={version}
if [ -z "$1" ]
then
	version=0
else
	version=$1
fi

if [ -z "$2" ]
then
	author='-dev-swlee'
else
	author=$2
fi

img=192.168.9.12:5000/hypercloud-api-server-test$author:$version
echo "Image = $img"
docker build -t $img  . 
docker push $img
docker rmi $img 
