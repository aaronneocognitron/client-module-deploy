#!/bin/bash

set -e

if [ $SUDO_USER ]; then user=$SUDO_USER; else user=`whoami`; fi

# check docker
if [ ! -x "$(command -v docker)" ]; then
  echo "Install docker..."
  source /etc/os-release

  if [ "$ID" != "ubuntu" ] && [ "$ID" != "debian" ]; then
      echo "Unsupported OS" 1>&2
      exit 1
  fi

  # install docker
  apt update -y
  apt install -y ca-certificates curl gnupg
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/$ID/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  chmod a+r /etc/apt/keyrings/docker.gpg
  echo \
    "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$ID \
    "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
    tee /etc/apt/sources.list.d/docker.list > /dev/null
  apt update -y
  apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

  usermod -aG docker $user
  service docker restart
fi

# check docker compose
if ! docker compose version &>/dev/null; then
  echo "Install docker compose..."
  apt update -y
  apt install -y docker-compose-plugin
fi