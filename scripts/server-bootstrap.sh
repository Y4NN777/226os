#!/usr/bin/env bash
# One-time preparation of a FRESH Ubuntu or Debian server for the 226os deploy workflow (docs/runbooks/deploy-setup.md).
# Run once by the owner, as root, on the server. Idempotent: running it again changes nothing that is already right.
#
# Usage: server-bootstrap.sh [--dry-run] [--yes] "<ssh public key of the CI deploy key>"
#   --dry-run  print the steps and change nothing (works without root)
#   --yes      do not ask before hardening sshd (needed without a terminal)
#
# What it does:
#   1. Docker Engine and the compose plugin, from Docker's official apt repository
#   2. make and curl (the deploy runs `make seed` and the smoke test on the server)
#   3. the `deploy` user, in the docker group, with ONLY the CI public key
#   4. ufw: deny incoming except 22, 80 and 443
#   5. unattended security upgrades
#   6. /opt/226os owned by deploy (env/, backups/, state/ inside)
#   7. sshd: no password login (a drop-in file, checked with sshd -t, then a reload: open sessions keep working).
#      Skipped, with a warning, when root has no authorized key, or when you do not confirm.
# Note: the `docker` group is root-equivalent; ufw does not filter ports that Docker publishes itself (here only
# Traefik publishes 80 and 443). Nothing prints the key material or any secret.
set -euo pipefail

usage() {
  printf 'usage: %s [--dry-run] [--yes] "<ssh public key>"\n' "$0" >&2
  exit 64
}
refuse() { printf 'REFUSED: %s\n' "$*" >&2; exit 2; }

dry=0
assume_yes=0
key=""
while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run) dry=1 ;;
    --yes) assume_yes=1 ;;
    -*) usage ;;
    *)
      [ -z "$key" ] || usage
      key="$1"
      ;;
  esac
  shift
done
[ -n "$key" ] || usage

# Arguments are checked before the root check: a wrong argument is a usage error whoever runs the script.
# The value is never echoed (a private key pasted by mistake must not reach a terminal log).
case "$key" in
  *$'\n'* | *$'\r'*) printf 'server-bootstrap: the key must be ONE line, an SSH public key\n' >&2; exit 64 ;;
esac
key_re='^(ssh-ed25519|ssh-rsa|ecdsa-sha2-nistp(256|384|521)|sk-ssh-ed25519@openssh\.com) [A-Za-z0-9+/=]+( [^[:cntrl:]]*)?$'
[[ "$key" =~ $key_re ]] || { printf 'server-bootstrap: the argument is not an SSH public key (ssh-ed25519 AAAA... comment)\n' >&2; exit 64; }

if [ "$dry" -eq 0 ] && [ "$(id -u)" -ne 0 ]; then
  refuse "run as root (this script installs packages and changes the firewall and sshd)"
fi

step_n=0
step() {
  step_n=$((step_n + 1))
  printf '\n[%s] %s\n' "$step_n" "$*"
}
[ "$dry" -eq 0 ] || printf 'server-bootstrap: dry run, nothing will be changed\n'

# --- 1. Docker ---------------------------------------------------------------------------------------------------
step "Docker Engine and the compose plugin (Docker's official apt repository)"
if [ "$dry" -eq 1 ]; then
  printf '    (dry run) install ca-certificates curl gnupg, add https://download.docker.com/linux/<distro>, install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin\n'
else
  # shellcheck disable=SC1091 # /etc/os-release exists on the target
  . /etc/os-release
  case "${ID-}" in ubuntu | debian) ;; *) refuse "only Ubuntu and Debian are supported (found '${ID-unknown}')" ;; esac
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -qq
  apt-get install -y -qq ca-certificates curl gnupg make
  install -m 0755 -d /etc/apt/keyrings
  if [ ! -s /etc/apt/keyrings/docker.asc ]; then
    curl -fsSL "https://download.docker.com/linux/$ID/gpg" -o /etc/apt/keyrings/docker.asc
    chmod a+r /etc/apt/keyrings/docker.asc
  fi
  arch="$(dpkg --print-architecture)"
  codename="${UBUNTU_CODENAME:-${VERSION_CODENAME:-}}"
  [ -n "$codename" ] || refuse "cannot find the distribution codename in /etc/os-release"
  printf 'deb [arch=%s signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/%s %s stable\n' "$arch" "$ID" "$codename" >/etc/apt/sources.list.d/docker.list
  apt-get update -qq
  apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  systemctl enable --now docker
  docker compose version
fi

# --- 2. deploy user -----------------------------------------------------------------------------------------------
step "the deploy user (docker group, the CI public key only)"
if [ "$dry" -eq 1 ]; then
  printf '    (dry run) create user deploy (no password), add it to the docker group, write ~deploy/.ssh/authorized_keys with the CI key\n'
else
  id deploy >/dev/null 2>&1 || adduser --disabled-password --gecos "" deploy
  usermod -aG docker deploy
  home="$(getent passwd deploy | cut -d: -f6)"
  install -d -m 700 -o deploy -g deploy "$home/.ssh"
  touch "$home/.ssh/authorized_keys"
  grep -qxF -- "$key" "$home/.ssh/authorized_keys" || printf '%s\n' "$key" >>"$home/.ssh/authorized_keys"
  chown deploy:deploy "$home/.ssh/authorized_keys"
  chmod 600 "$home/.ssh/authorized_keys"
fi

# --- 3. firewall ----------------------------------------------------------------------------------------------------
step "ufw: deny incoming, allow 22, 80 and 443 only"
if [ "$dry" -eq 1 ]; then
  printf '    (dry run) install ufw; default deny incoming, allow outgoing; allow 22/tcp 80/tcp 443/tcp; enable\n'
else
  apt-get install -y -qq ufw
  ufw default deny incoming
  ufw default allow outgoing
  # SSH first: enabling the firewall must never cut the session this script runs in.
  ufw allow 22/tcp
  ufw allow 80/tcp
  ufw allow 443/tcp
  ufw --force enable
fi

# --- 4. unattended upgrades -------------------------------------------------------------------------------------------
step "unattended-upgrades (security updates)"
if [ "$dry" -eq 1 ]; then
  printf '    (dry run) install unattended-upgrades and enable the daily run\n'
else
  apt-get install -y -qq unattended-upgrades
  printf 'APT::Periodic::Update-Package-Lists "1";\nAPT::Periodic::Unattended-Upgrade "1";\n' >/etc/apt/apt.conf.d/20auto-upgrades
fi

# --- 5. /opt/226os --------------------------------------------------------------------------------------------------------
step "/opt/226os owned by deploy (env, backups, state)"
if [ "$dry" -eq 1 ]; then
  printf '    (dry run) create /opt/226os/{env,backups,state} (mode 700), owner deploy\n'
else
  install -d -m 750 -o deploy -g deploy /opt/226os
  install -d -m 700 -o deploy -g deploy /opt/226os/env /opt/226os/backups /opt/226os/state
fi

# --- 6. sshd -----------------------------------------------------------------------------------------------------------------
step "sshd: no password login"
if [ "$dry" -eq 1 ]; then
  printf '    (dry run) warn, then write /etc/ssh/sshd_config.d/50-226os.conf (PasswordAuthentication no), sshd -t, reload sshd\n'
else
  echo "    WARNING: after this, only SSH keys can log in. Keep THIS session open and test a key login from another terminal."
  if [ ! -s /root/.ssh/authorized_keys ]; then
    printf '    SKIPPED: /root/.ssh/authorized_keys is empty or missing: disabling passwords could lock you out. Add your key, then run this script again.\n'
  else
    answer=no
    if [ "$assume_yes" -eq 1 ]; then
      answer=yes
    elif [ -t 0 ]; then
      read -r -p "    Disable SSH password login now? [yes/no] " answer
    else
      printf '    SKIPPED: no terminal to confirm; run again with --yes to disable password login.\n'
    fi
    if [ "$answer" = yes ]; then
      install -d -m 755 /etc/ssh/sshd_config.d
      printf 'PasswordAuthentication no\nKbdInteractiveAuthentication no\nPermitRootLogin prohibit-password\n' >/etc/ssh/sshd_config.d/50-226os.conf
      if sshd -t; then
        systemctl reload ssh 2>/dev/null || systemctl reload sshd
        echo "    password login disabled (sshd reloaded, existing sessions kept)"
      else
        rm -f /etc/ssh/sshd_config.d/50-226os.conf
        refuse "sshd -t rejected the configuration: the change was removed, sshd is untouched"
      fi
    else
      printf '    SKIPPED: not confirmed.\n'
    fi
  fi
fi

# --- end -------------------------------------------------------------------------------------------------------------------------
printf '\nDone. Next, from your machine:\n'
printf '  1. Host key for the DEPLOY_KNOWN_HOSTS secret (public, pin it so CI never trusts a first connection):\n'
if [ "$dry" -eq 1 ]; then
  printf '       <DEPLOY_HOST> <the ssh-ed25519 line of /etc/ssh/ssh_host_ed25519_key.pub>\n'
elif [ -r /etc/ssh/ssh_host_ed25519_key.pub ]; then
  printf '       <DEPLOY_HOST> %s\n' "$(cut -d' ' -f1,2 /etc/ssh/ssh_host_ed25519_key.pub)"
fi
printf '     (replace <DEPLOY_HOST> by the exact value of the DEPLOY_HOST variable)\n'
printf '  2. Then follow docs/runbooks/deploy-setup.md (secrets, variables, DNS).\n'
