#!/usr/bin/bash

function itoa {
    ip=${1}
    if [[ $ip -lt 0 ]] ; then
        ip=$((2147483647+$((2147483648+${ip}))))
        echo -n "$ip -> "
    fi
    #returns the dotted-decimal ascii form of an IP arg passed in integer format
    echo -n $(($(($(($((${ip}/256))/256))/256))%256)).
    echo -n $(($(($((${ip}/256))/256))%256)).
    echo -n $(($((${ip}/256))%256)).
    echo "$((${ip}%256))"
}

jq --version || sudo apt install -y jq

config_url="${1:-https://ton-blockchain.github.io/testnet-global.config.json}"
echo "- Using config URL: $config_url"
IFS_save=$IFS
IPs="$(curl "$config_url" | jq '.liteservers[].ip')"
for ip in $(IFS="\n" echo $IPs) ; do
    IP="$(itoa $ip)"
    echo "-- $ip -> $IP"
    ping -c 3 "$IP"
done
IFS=$IFS_save