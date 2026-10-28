#!/usr/bin/env bash
# up: запускает ВМ, ждёт готовности k3s и печатает IP. down: останавливает ВМ.
# Нужны yc с ролью compute.operator на ВМ, jq и для up SSH-ключ (SSH_KEY).
set -euo pipefail

VM_ID=fv4v4f590n4gbbdd607m
SSH_USER=shugraph
SSH_KEY=${SSH_KEY:-$HOME/.ssh/shugraph-ci}

vm() { yc compute instance get --id "$VM_ID" --format json; }
status() { vm | jq -r .status; }

# Во время сбоев облака yc может упасть, хотя операция прошла, поэтому ориентируемся на статус ВМ, а не на код возврата.
wait_status() {
  for _ in $(seq 60); do
    [[ $(status) == "$1" ]] && return
    sleep 5
  done
  echo "ВМ не перешла в $1 за 5 минут, статус: $(status)" >&2
  return 1
}

up() {
  case $(status) in
    STOPPING) wait_status STOPPED; yc compute instance start --id "$VM_ID" --async || true ;;
    STOPPED) yc compute instance start --id "$VM_ID" --async || true ;;
  esac
  wait_status RUNNING

  local ip err
  ip=$(vm | jq -r '.network_interfaces[0].primary_v4_address.one_to_one_nat.address')
  local ssh=(ssh -i "$SSH_KEY" -o BatchMode=yes -o StrictHostKeyChecking=accept-new -o ConnectTimeout=10 "$SSH_USER@$ip")
  for _ in $(seq 30); do
    if err=$("${ssh[@]}" 'sudo k3s kubectl wait --for=condition=Ready node --all --timeout=10s' 2>&1); then
      "${ssh[@]}" 'sudo k3s kubectl get nodes'
      echo "IP: $ip"
      [[ -n ${GITHUB_STEP_SUMMARY:-} ]] && echo "IP ВМ: \`$ip\`" >> "$GITHUB_STEP_SUMMARY"
      return 0
    fi
    sleep 10
  done
  echo "k3s не стал Ready за 10 минут, последняя попытка: $err" >&2
  return 1
}

down() {
  case $(status) in
    STARTING) wait_status RUNNING; yc compute instance stop --id "$VM_ID" --async || true ;;
    RUNNING) yc compute instance stop --id "$VM_ID" --async || true ;;
  esac
  wait_status STOPPED
}

case ${1:-} in
  up | down) "$1" ;;
  *) echo "usage: $0 up|down" >&2; exit 1 ;;
esac
