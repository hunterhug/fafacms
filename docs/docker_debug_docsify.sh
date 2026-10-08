#!/bin/bash
docker run -it --rm -p 9889:3000 --name=docsify -v $(pwd):/docs hunterhug/docsify