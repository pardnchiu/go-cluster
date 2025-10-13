#!/bin/bash

get_health() {
  local hostname=$(hostname)
  
  # disk usage get bytes
  # /dev/sda1      8242348032 1658474496 6198059008  22% /
  local disk_info=$(df -B1 / | tail -1)
  local disk_max=$(echo "$disk_info" | awk '{print $2}')
  local disk=$(echo "$disk_info" | awk '{print $3}')
  
  # cpu usage
  read user nice system idle iowait irq softirq < <(grep '^cpu ' /proc/stat | awk '{print $2, $3, $4, $5, $6, $7, $8}')
  local total1=$((user + nice + system + idle + iowait + irq + softirq))
  local idle1=$idle
  
  # delay 1 second to get this 1 second cpu usage
  sleep 1
  
  read user nice system idle iowait irq softirq < <(grep '^cpu ' /proc/stat | awk '{print $2, $3, $4, $5, $6, $7, $8}')
  local total2=$((user + nice + system + idle + iowait + irq + softirq))
  local idle2=$idle
  
  local total=$((total2 - total1))
  local idle=$((idle2 - idle1))
  # just need **.**%
  local cpu=$(awk "BEGIN {printf \"%.4f\", (($total - $idle) / $total)}")
  local cpu_max=$(nproc)
  
  # memory usage get bytes
  # Mem:      2062483456  1114624000   386306048      569344   718159872   947859456
  local mem_info=$(free -b | grep Mem)
  local mem_max=$(echo "$mem_info" | awk '{print $2}')
  local mem=$(echo "$mem_info" | awk '{print $3}')
  
  # uptime
  # 2415.85 4766.80
  local uptime_sec=$(awk '{print int($1)}' /proc/uptime)
  local status="online"
  
  cat <<EOF
{
  "cpu": $cpu,
  "disk": $disk,
  "id": "node/$hostname",
  "maxcpu": $cpu_max,
  "maxdisk": $disk_max,
  "maxmem": $mem_max,
  "mem": $mem,
  "node": "$hostname",
  "status": "$status",
  "uptime": $uptime_sec
}
EOF
}

get_health