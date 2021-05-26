#!/bin/sh
./gobuild.sh
img=192.168.6.110:5000/hypercloud-api-server:v9.0
docker rmi $img 
docker build -t $img  . 
docker push $img 
