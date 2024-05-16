# create /dev/null
mknod /dev/null c 1 3
chmod 666 /dev/null

chmod +x /mnt/nimono
cd /mnt
timeout --signal=2 1 ./nimono
