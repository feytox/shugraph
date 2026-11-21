#!/usr/bin/env bash
# Включает и выключает ВМ с k3s.
#   up    запускает ВМ, ждёт готовности k3s и печатает IP
#   down  останавливает ВМ
# Нужны yc с ролью compute.operator на ВМ и jq, для up ещё ssh.
# VM_ID: ID ВМ. Если не задан, скрипт берёт переменную VM_ID окружения vm на GitHub через gh.
# SSH_KEY: ключ для up, по умолчанию ~/.ssh/shugraph-ci.
set -euo pipefail

VM_ID=${VM_ID:-$(gh variable get VM_ID --env vm)}
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

SSHD_WAIT=120
SSHD_CHECKS=5
MAX_REISSUES=3

start() {
  for _ in $(seq 12); do
    yc compute instance start --id "$VM_ID" --async > /dev/null && return
    [[ $(status) != STOPPED ]] && return
    sleep 5
  done
  echo "ВМ не запустилась за минуту, статус: $(status)" >&2
  return 1
}

stop() {
  yc compute instance stop --id "$VM_ID" --async > /dev/null || true
  wait_status STOPPED
}

address() { vm | jq -r '.network_interfaces[0].primary_v4_address.one_to_one_nat.address'; }

sshd_stable() {
  local ok=0 deadline=$((SECONDS + SSHD_WAIT))
  while (( SECONDS < deadline )); do
    if [[ -n $(ssh-keyscan -T 5 -t ed25519 "$1" 2> /dev/null) ]]; then
      (( ++ok == SSHD_CHECKS )) && return
    else
      ok=0
    fi
    sleep 1
  done
  return 1
}

wait_k3s() {
  local ip=$1 err rc
  local ssh=(timeout 30 ssh -i "$SSH_KEY" -o BatchMode=yes -o StrictHostKeyChecking=accept-new -o ConnectTimeout=10
    -o ServerAliveInterval=5 -o ServerAliveCountMax=2 "$SSH_USER@$ip")
  for _ in $(seq 30); do
    rc=0
    err=$("${ssh[@]}" 'sudo k3s kubectl wait --for=condition=Ready node --all --timeout=10s' 2>&1) || rc=$?
    if (( rc == 0 )); then
      "${ssh[@]}" 'sudo k3s kubectl get nodes'
      return 0
    fi
    (( rc == 124 )) && err="ssh не ответил за 30 секунд"
    echo "$(date -u +%T) k3s на $ip не готов: ${err##*$'\n'}" >&2
    sleep 10
  done
  echo "k3s не стал Ready за 10 минут" >&2
  return 1
}

up() {
  case $(status) in
    STOPPING) wait_status STOPPED; start ;;
    STOPPED) start ;;
  esac
  wait_status RUNNING

  local ip reissue
  for reissue in $(seq 0 "$MAX_REISSUES"); do
    ip=$(address)
    sshd_stable "$ip" && break
    if (( reissue == MAX_REISSUES )); then
      echo "SSH на ВМ недоступен после $MAX_REISSUES перезапусков, последний IP $ip" >&2
      return 1
    fi
    echo "$(date -u +%T) IP $ip недоступен: sshd не ответил $SSHD_CHECKS раз подряд за $SSHD_WAIT с, перезапускаю ВМ" >&2
    stop
    start
    wait_status RUNNING
  done

  wait_k3s "$ip"
  echo "IP: $ip"
  [[ -n ${GITHUB_STEP_SUMMARY:-} ]] && echo "IP ВМ: \`$ip\`" >> "$GITHUB_STEP_SUMMARY"
  [[ -n ${GITHUB_OUTPUT:-} ]] && echo "ip=$ip" >> "$GITHUB_OUTPUT"
  return 0
}

down() {
  case $(status) in
    STARTING) wait_status RUNNING; stop ;;
    RUNNING) stop ;;
    *) wait_status STOPPED ;;
  esac
}

case ${1:-} in
  up | down) "$1" ;;
  *) echo "usage: $0 up|down" >&2; exit 1 ;;
esac
