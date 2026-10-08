#!/bin/bash
docker rm -f fafadoc
docker run --name fafadoc -d -p 9889:80 hunterhug/fafadoc