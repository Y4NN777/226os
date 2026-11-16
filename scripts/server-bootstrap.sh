#!/usr/bin/env bash
# One-time preparation of a FRESH Ubuntu or Debian server for the 226os deploy workflow (docs/runbooks/deploy-setup.md).
# Run once by the owner, as root, on the server. Idempotent: running it again changes nothing that is already right.
#
# Usage: server-bootstrap.sh [--dry-run] [--ssh-fingerprint SHA256:...] "<ssh public key of the CI deploy key>"
#   --dry-run                      print the steps and change nothing (works without root)
#   --ssh-fingerprint SHA256:...   the fingerprint of YOUR key that logs in as root (`ssh-keygen -lf` of your public
#                                  key). Without it (or an interactive answer naming it) sshd is not touched.
#
# What it does:
#   1. Docker Engine and the compose plugin, from Docker's official apt repository
#   2. make, curl and age (the deploy runs `make seed`, the smoke test and the backup encryption on the server) (the deploy runs `make seed` and the smoke test on the server)
#   3. the `deploy` user, in the docker group, with ONLY the CI public key
#   4. ufw: deny incoming except the sshd port (rate limited), 80 and 443
#   5. unattended security upgrades
#   6. /opt/226os owned by deploy (env/, backups/, state/ inside)
#   7. sshd: no password login (a drop-in file, checked with sshd -t, then a reload: open sessions keep working).
#      Skipped, with a warning, unless the fingerprint you give is one of the keys in /root/.ssh/authorized_keys.
#      After the reload `sshd -T` must report passwordauthentication no, or the change is rolled back.
# Note: the `docker` group is root-equivalent; ufw does not filter ports that Docker publishes itself (here only
# Traefik publishes 80 and 443). Nothing prints the key material or any secret.
set -euo pipefail

usage() {
  printf 'usage: %s [--dry-run] [--ssh-fingerprint SHA256:...] "<ssh public key>"\n' "$0" >&2
  exit 64
}
refuse() { printf 'REFUSED: %s\n' "$*" >&2; exit 2; }

dry=0
fingerprint=""
key=""
while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run) dry=1 ;;
    --ssh-fingerprint)
      [ $# -ge 2 ] || usage
      fingerprint="$2"
      shift
      ;;
    -*) usage ;;
    *)
      [ -z "$key" ] || usage
      key="$1"
      ;;
  esac
  shift
done
[ -n "$key" ] || usage
[ -z "$fingerprint" ] || [[ "$fingerprint" =~ ^SHA256:[A-Za-z0-9+/]{43}$ ]] || { printf 'server-bootstrap: --ssh-fingerprint must look like SHA256:<43 characters> (ssh-keygen -lf of your public key)\n' >&2; exit 64; }

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
  printf '    (dry run) install ca-certificates curl gnupg make age, verify the pinned Docker key fingerprint, add https://download.docker.com/linux/<distro>, install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin\n'
else
  # shellcheck disable=SC1091 # /etc/os-release exists on the target
  . /etc/os-release
  case "${ID-}" in ubuntu | debian) ;; *) refuse "only Ubuntu and Debian are supported (found '${ID-unknown}')" ;; esac
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -qq
  apt-get install -y -qq ca-certificates curl gnupg make age
  install -m 0755 -d /etc/apt/keyrings
  if [ ! -s /etc/apt/keyrings/docker.asc ]; then
    # The key is pinned: it is used only if its fingerprint is Docker's published one.
    docker_fpr="9DC858229FC7DD38854AE2D88D81803C0EBFCD88"
    keytmp="$(mktemp)"
    curl -fsSL "https://download.docker.com/linux/$ID/gpg" -o "$keytmp"
    gpg --show-keys --with-colons "$keytmp" | awk -F: '/^fpr:/ {print $10}' | grep -qx "$docker_fpr" ||
      { rm -f "$keytmp"; refuse "the Docker apt key does not have the pinned fingerprint $docker_fpr: nothing installed"; }
    install -m 0644 "$keytmp" /etc/apt/keyrings/docker.asc
    rm -f "$keytmp"
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
  printf '    (dry run) create user deploy (no password), add it to the docker group, write ~deploy/.ssh/authorized_keys with the CI key as "restrict <key>"\n'
else
  id deploy >/dev/null 2>&1 || adduser --disabled-password --gecos "" deploy
  usermod -aG docker deploy
  home="$(getent passwd deploy | cut -d: -f6)"
  install -d -m 700 -o deploy -g deploy "$home/.ssh"
  touch "$home/.ssh/authorized_keys"
  # `restrict`: no port forwarding, agent forwarding, X11 or pty for the CI key (it only runs commands).
  grep -qxF -- "restrict $key" "$home/.ssh/authorized_keys" || printf 'restrict %s\n' "$key" >>"$home/.ssh/authorized_keys"
  chown deploy:deploy "$home/.ssh/authorized_keys"
  chmod 600 "$home/.ssh/authorized_keys"
fi

# --- 3. firewall ----------------------------------------------------------------------------------------------------
step "ufw: deny incoming, allow the sshd port (limited), 80 and 443 only"
if [ "$dry" -eq 1 ]; then
  printf '    (dry run) install ufw; default deny incoming, allow outgoing; read the sshd port from sshd -T and `ufw limit` it; allow 80/tcp 443/tcp; enable\n'
else
  apt-get install -y -qq ufw
  ufw default deny incoming
  ufw default allow outgoing
  # SSH first: enabling the firewall must never cut the session this script runs in. The port is the one sshd really
  # listens on (not an assumed 22); `limit` also rate-limits repeated connection attempts.
  sshd_conf="$(sshd -T 2>/dev/null)" || refuse "sshd -T failed: the firewall was not enabled"
  ssh_port="$(awk '$1 == "port" {print $2; exit}' <<<"$sshd_conf")"
  [[ "$ssh_port" =~ ^[0-9]+$ ]] || refuse "cannot read the sshd port from sshd -T: the firewall was not enabled"
  # An older plain `allow 22` would keep port 22 open without the rate limit (and when sshd listens elsewhere).
  ufw delete allow 22/tcp >/dev/null 2>&1 || true
  ufw limit "$ssh_port/tcp"
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
  printf '    (dry run) warn; require --ssh-fingerprint (or the typed answer) to be a key of /root/.ssh/authorized_keys; write /etc/ssh/sshd_config.d/50-226os.conf (PasswordAuthentication no), sshd -t, reload sshd, verify sshd -T, roll back if it does not say passwordauthentication no\n'
else
  echo "    WARNING: after this, only SSH keys can log in. Keep THIS session open and test a key login from another terminal."
  if [ ! -s /root/.ssh/authorized_keys ]; then
    printf '    SKIPPED: /root/.ssh/authorized_keys is empty or missing: disabling passwords could lock you out. Add your key, then run this script again.\n'
  else
    fp="$fingerprint"
    if [ -z "$fp" ] && [ -t 0 ]; then
      echo "    Keys that can log in as root:"
      ssh-keygen -lf /root/.ssh/authorized_keys | sed 's/^/      /'
      read -r -p "    Type the SHA256 fingerprint of the key YOU use, to disable password login (empty to skip): " fp
    fi
    if [ -z "$fp" ]; then
      printf '    SKIPPED: no fingerprint given (run again with --ssh-fingerprint SHA256:...).\n'
    elif ! ssh-keygen -lf /root/.ssh/authorized_keys | awk '{print $2}' | grep -qxF -- "$fp"; then
      printf '    SKIPPED: that fingerprint is not a key of /root/.ssh/authorized_keys: password login stays enabled.\n'
    else
      install -d -m 755 /etc/ssh/sshd_config.d
      printf 'PasswordAuthentication no\nKbdInteractiveAuthentication no\nPermitRootLogin prohibit-password\n' >/etc/ssh/sshd_config.d/50-226os.conf
      reload_sshd() { systemctl reload ssh 2>/dev/null || systemctl reload sshd; }
      if sshd -t; then
        reload_sshd
        sshd_after="$(sshd -T)"
        if grep -qi '^passwordauthentication no' <<<"$sshd_after"; then
          echo "    password login disabled (sshd reloaded, existing sessions kept, verified with sshd -T)"
        else
          rm -f /etc/ssh/sshd_config.d/50-226os.conf
          reload_sshd
          refuse "sshd -T does not report passwordauthentication no after the reload (another file wins?): the change was removed"
        fi
      else
        rm -f /etc/ssh/sshd_config.d/50-226os.conf
        refuse "sshd -t rejected the configuration: the change was removed, sshd is untouched"
      fi
    fi
  fi
fi

# --- end -------------------------------------------------------------------------------------------------------------------------
printf '\nDone. Next, from your machine:\n'
printf '  1. Host key line for DEPLOY_KNOWN_HOSTS in deploy/<staging|production>.conf (public, pin it so CI never trusts a first connection):\n'
if [ "$dry" -eq 1 ]; then
  printf '       <DEPLOY_HOST> <the ssh-ed25519 line of /etc/ssh/ssh_host_ed25519_key.pub>\n'
elif [ -r /etc/ssh/ssh_host_ed25519_key.pub ]; then
  printf '       <DEPLOY_HOST> %s\n' "$(cut -d' ' -f1,2 /etc/ssh/ssh_host_ed25519_key.pub)"
fi
printf '     (replace <DEPLOY_HOST> by the exact value of DEPLOY_HOST in the same deploy/<env>.conf, then open a pull request)\n'
printf '  2. Then follow docs/runbooks/deploy-setup.md (conf file, the two environment secrets, DNS).\n'
