#!/bin/bash

if [ "$#" -lt 1 ]; then
        echo "Please provide dir to create the root file system in"
        exit 1
fi


ROOT=$1

if [ -e "$ROOT" ]
then
        echo "$ROOT already exists"
        exit 1
fi

mkdir -p "$ROOT"/{etc,tmp,proc,sys,dev,home,mnt,root,usr/{bin,sbin,lib},var} && chmod a+rwxt "$ROOT"/tmp

cat > "$ROOT"/etc/passwd << 'EOF' &&
root::0:0:root:/root:/bin/sh
guest:x:500:500:guest:/home/guest:/bin/sh
nobody:x:65534:65534:nobody:/proc/self:/dev/null
EOF

cat > "$ROOT"/etc/group << 'EOF' &&
root:x:0:
guest:x:500:
EOF

echo "nameserver 8.8.8.8" > "$ROOT"/etc/resolv.conf


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-extldflags '-static'" -o "$ROOT"/init main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-extldflags '-static'" -o "$ROOT"/hello hello.go

cd "$ROOT"

find . | cpio -o -H newc | gzip > ../root.cpio.gz
cd ..
echo "root.cpio.gz created"

