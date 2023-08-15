#!/bin/bash
set -e

configPath=$1
if [ $SUDO_USER ]; then user=$SUDO_USER; else user=`whoami`; fi

if [[ ! -e $configPath/docker-compose.yml ]]; then
    touch $configPath/docker-compose.yml
    chown "$user:$user" $configPath/docker-compose.yml
fi