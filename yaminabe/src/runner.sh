# create /dev/null
mknod /dev/null c 1 3
chmod 666 /dev/null

mkdir -p /home/root
echo "hello, world!1" > /home/root/hello1.txt
echo "hello, world!2" > /home/root/hello2.txt
echo "hello, world!3" > /home/root/hello3.txt
echo "hello, world!4" > /home/root/hello4.txt

chmod +x /mnt/nimono
/mnt/nimono &
nimono_pid=$!
sleep 1
/mnt/target &
target_pid=$!
wait $target_pid

kill -2 $nimono_pid
wait $nimono_pid
cp ./syscall_events.json /mnt/
echo $target_pid > /mnt/target_pid
