#!/bin/bash 

redisip=192.168.104.55
redisport=6379
redispasswd=zjxl2018


echo "begin..."
echo "first get data from redis"
redis-cli -h  $redisip -p $redisport -a $redispasswd hgetall lbs.phone  > a

echo "filter sim"
awk '++i%2' a  > b

echo "create sim.txt"
> sim.txt
cat  b  >> sim.txt

echo "rm tmp file"
rm -f a b 
echo "finish..."
