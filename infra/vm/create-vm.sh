#!/usr/bin/env bash
# Создаёт и запускает прерываемую ВМ с k3s в каталоге из текущего профиля yc. Запускается один раз вручную.
# Нужны yc и jq, в каталоге сеть default с подсетью default-<зона>.
set -euo pipefail
cd "$(dirname "$0")"

ZONE=ru-central1-d

for cmd in yc jq; do
  command -v "$cmd" > /dev/null || { echo "Не найден $cmd" >&2; exit 1; }
done
yc vpc subnet get "default-$ZONE" > /dev/null

SG_ID=$(yc vpc security-group create \
  --name shucluster-group \
  --network-name default \
  --rule "direction=ingress,port=22,protocol=tcp,v4-cidrs=[0.0.0.0/0]" \
  --rule "direction=ingress,port=80,protocol=tcp,v4-cidrs=[0.0.0.0/0]" \
  --rule "direction=egress,port=any,protocol=any,v4-cidrs=[0.0.0.0/0]" \
  --format json | jq -r .id)

yc compute instance create \
  --name shucluster-vm \
  --hostname shucluster-vm \
  --zone "$ZONE" \
  --platform standard-v3 \
  --cores 2 \
  --memory 4 \
  --core-fraction 20 \
  --preemptible \
  --create-boot-disk name=shucluster-disk,type=network-hdd,size=20,image-folder-id=standard-images,image-family=ubuntu-2404-lts,auto-delete=yes \
  --network-interface subnet-name="default-$ZONE",security-group-ids="$SG_ID",nat-ip-version=ipv4 \
  --metadata-from-file user-data=cloud-init.yaml
