# create /dev/null
mknod /dev/null c 1 3
chmod 666 /dev/null

chmod +x /mnt/nimono
/mnt/nimono &
nimono_pid=$!
sleep 1
/mnt/target

kill -2 $nimono_pid
wait $nimono_pid
cp ./syscall_events.json /mnt/
