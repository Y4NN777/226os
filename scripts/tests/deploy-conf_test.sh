#!/usr/bin/env bash
# Tests for scripts/deploy-conf.sh (the non-secret deploy settings of an environment, read from the commit being
# deployed) and for the way deploy.yml uses it. Nothing touches a server or the network.
# Usage: scripts/tests/deploy-conf_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
conf="$root/scripts/deploy-conf.sh"
wf="$root/.github/workflows/deploy.yml"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

kh='203.0.113.7 ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl'
valid() {
  printf '# comment\n\nDEPLOY_HOST=203.0.113.7\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=%s\n' "$kh"
}

# run NAME CONTENT: the parser on a file with this content (printf %b of CONTENT); sets rc, out, err.
run() {
  printf '%b' "$2" >"$tmp/$1.conf"
  rc=0
  out="$("$conf" staging "$tmp/$1.conf" 2>"$tmp/err")" || rc=$?
  err="$(cat "$tmp/err")"
}
# refused NAME CONTENT LABEL: refused (exit 2), nothing on stdout, and the offending value is not echoed in clear.
refused() {
  run "$1" "$2"
  if [ "$rc" -eq 2 ] && [ -z "$out" ]; then ok "$3"; else ko "$3 (rc=$rc out=$out err=$err)"; fi
}

[ -x "$conf" ] || { ko "scripts/deploy-conf.sh is missing or not executable"; finish; exit $?; }

run valid "$(valid)\n"
want="DEPLOY_HOST=203.0.113.7
DEPLOY_USER=deploy
PUBLIC_HOST=staging.226os.tech
DEPLOY_KNOWN_HOSTS=$kh"
[ "$rc" -eq 0 ] && [ "$out" = "$want" ] && ok "a valid file prints the four validated KEY=value lines" || ko "valid (rc=$rc out=$out err=$err)"

# No trailing newline at the end of the file is fine.
run nonl "$(valid)"
[ "$rc" -eq 0 ] && ok "a last line without newline is read" || ko "no trailing newline (rc=$rc err=$err)"

# A name instead of an IP, a DNS host key line.
run dnshost "DEPLOY_HOST=srv-1.example.test\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=srv-1.example.test ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMq\n"
[ "$rc" -eq 0 ] && ok "a DNS name works as DEPLOY_HOST" || ko "dns host (rc=$rc err=$err)"

# The profile argument and the file.
rc=0
"$conf" staging "$tmp/nope.conf" >/dev/null 2>&1 || rc=$?
[ "$rc" -eq 2 ] && ok "a missing file is refused" || ko "missing file (rc=$rc)"
rc=0
"$conf" prod "$tmp/valid.conf" >/dev/null 2>&1 || rc=$?
[ "$rc" -ne 0 ] && ok "an unknown target is refused" || ko "unknown target (rc=$rc)"
rc=0
"$conf" >/dev/null 2>&1 || rc=$?
[ "$rc" -ne 0 ] && ok "no target is refused" || ko "no target (rc=$rc)"

# Each field invalid, with the same regexes as deploy.yml had.
refused host_space "DEPLOY_HOST=a b\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "DEPLOY_HOST with a space is refused"
refused host_slash "DEPLOY_HOST=a/b\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "DEPLOY_HOST with a slash is refused"
refused host_empty "DEPLOY_HOST=\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "an empty DEPLOY_HOST is refused"
refused user_upper "DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=Deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "an upper-case DEPLOY_USER is refused"
refused user_at "DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=root@evil\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "a DEPLOY_USER with @ is refused"
refused pub_upper "DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=deploy\nPUBLIC_HOST=Staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "an upper-case PUBLIC_HOST is refused"
refused pub_nodot "DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=deploy\nPUBLIC_HOST=localhost\nDEPLOY_KNOWN_HOSTS=$kh\n" "a PUBLIC_HOST that is not a DNS name is refused"
refused pub_ip "DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=deploy\nPUBLIC_HOST=203.0.113.7\nDEPLOY_KNOWN_HOSTS=$kh\n" "a PUBLIC_HOST that is an IP address is refused"
refused kh_type "DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=203.0.113.7 ssh-dss AAAA\n" "a known_hosts line with another key type is refused"
refused kh_other_host "DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=203.0.113.8 ssh-ed25519 AAAAC3Nza\n" "a known_hosts line for another host than DEPLOY_HOST is refused"
refused kh_wild "DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=* ssh-ed25519 AAAAC3Nza\n" "a wildcard known_hosts line is refused"
refused kh_extra "DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh extra\n" "a known_hosts line with a trailing field is refused"
refused kh_marker "DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=@cert-authority 203.0.113.7 ssh-ed25519 AAAAC3Nza\n" "a known_hosts marker is refused"

# Keys: missing, unknown, duplicate, lower case, spaces around =.
refused missing "DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\n" "a missing key is refused"
refused unknown "$(valid)\nEXTRA=1\n" "an unknown key is refused"
refused dup "$(valid)\nDEPLOY_USER=other\n" "a duplicate key is refused"
refused lower "$(valid)\ndeploy_user=x\n" "a lower-case key is refused"
refused spaced "DEPLOY_HOST = 203.0.113.7\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "spaces around = are refused"
refused noeq "$(valid)\nJUNKLINE\n" "a line that is not KEY=value is refused"
refused indented "$(valid)\n DEPLOY_USER=deploy\n" "an indented line is refused"
refused export "export DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "an export line is refused"
refused empty "" "an empty file is refused"

# Placeholders: refused with a message that names the file to fill.
printf 'DEPLOY_HOST=FILL-ME\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=FILL-ME ssh-ed25519 FILL-ME\n' >"$tmp/ph.conf"
rc=0
out="$("$conf" staging "$tmp/ph.conf" 2>&1)" || rc=$?
[ "$rc" -eq 2 ] && grep -q 'fill deploy/staging.conf' <<<"$out" && ok "a placeholder is refused and says: fill deploy/staging.conf" || ko "placeholder (rc=$rc out=$out)"
printf 'DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=deploy\nPUBLIC_HOST=226os.tech\nDEPLOY_KNOWN_HOSTS=FILL-ME ssh-ed25519 FILL-ME\n' >"$tmp/ph2.conf"
rc=0
out="$("$conf" production "$tmp/ph2.conf" 2>&1)" || rc=$?
[ "$rc" -eq 2 ] && grep -q 'fill deploy/production.conf' <<<"$out" && ok "a placeholder in one value only is refused (production)" || ko "placeholder production (rc=$rc out=$out)"

# Injection attempts: nothing on stdout, exit 2, the payload is never executed (marker file stays absent).
refused inj_sub "DEPLOY_HOST=\$(touch $tmp/pwned)\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "a command substitution in a value is refused"
refused inj_tick "DEPLOY_HOST=\`touch $tmp/pwned\`\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "a backtick in a value is refused"
refused inj_semi "DEPLOY_HOST=203.0.113.7;touch $tmp/pwned\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "a semicolon in a value is refused"
refused inj_amp "DEPLOY_HOST=203.0.113.7&&id\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "an ampersand in a value is refused"
refused inj_pipe "DEPLOY_USER=deploy|id\nDEPLOY_HOST=203.0.113.7\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "a pipe in a value is refused"
refused inj_dq "DEPLOY_HOST=\"203.0.113.7\"\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "double quotes around a value are refused"
refused inj_sq "DEPLOY_HOST='203.0.113.7'\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "single quotes around a value are refused"
refused inj_dollar "DEPLOY_HOST=\${HOME}\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "a variable reference in a value is refused"
refused inj_cr "DEPLOY_HOST=203.0.113.7\r\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "a carriage return is refused"
refused inj_nl_kh "DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n$kh\n" "a second host key line (a repeated key) is refused"
refused inj_newline_env "DEPLOY_HOST=203.0.113.7\nGITHUB_ENV=/tmp/x\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "a line that would set another variable is refused"
refused inj_tab "DEPLOY_HOST=203.0.113.7\tx\nDEPLOY_USER=deploy\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n" "a tab in a value is refused"
[ ! -e "$tmp/pwned" ] && ok "no payload ran" || ko "a payload ran"

# Errors name the key and never print the value.
run secretish "DEPLOY_HOST=203.0.113.7\nDEPLOY_USER=Sup3rSecretValue\nPUBLIC_HOST=staging.226os.tech\nDEPLOY_KNOWN_HOSTS=$kh\n"
[ "$rc" -eq 2 ] && grep -q DEPLOY_USER <<<"$err" && ! grep -q Sup3rSecretValue <<<"$err" && ok "an error names the key, not the value" || ko "error text (rc=$rc err=$err)"

# The committed files: right keys, and refused while they hold a placeholder (the deploy job must say so).
for t in staging production; do
  f="$root/deploy/$t.conf"
  [ -f "$f" ] && ok "deploy/$t.conf is committed" || { ko "deploy/$t.conf is missing"; continue; }
  rc=0
  out="$("$conf" "$t" "$f" 2>&1)" || rc=$?
  if [ "$rc" -eq 0 ]; then ok "deploy/$t.conf is filled and valid"; elif grep -q "fill deploy/$t.conf" <<<"$out"; then ok "deploy/$t.conf is a placeholder file and the parser says: fill deploy/$t.conf"; else ko "deploy/$t.conf (rc=$rc out=$out)"; fi
  grep -qxE 'DEPLOY_USER=deploy' "$f" && grep -qE '^PUBLIC_HOST=' "$f" && grep -qE '^DEPLOY_HOST=' "$f" && grep -qE '^DEPLOY_KNOWN_HOSTS=' "$f" && ok "deploy/$t.conf has the four keys" || ko "deploy/$t.conf keys"
done
grep -qx 'PUBLIC_HOST=staging.226os.tech' "$root/deploy/staging.conf" && grep -qx 'PUBLIC_HOST=226os.tech' "$root/deploy/production.conf" && ok "the public hosts are the ones of the encrypted files" || ko "public hosts"

# deploy.yml: the settings come from the file, only two secrets remain, no fallback to the old sources.
for pat in 'vars\.DEPLOY_HOST' 'vars\.DEPLOY_USER' 'vars\.PUBLIC_HOST' 'secrets\.DEPLOY_KNOWN_HOSTS' 'secrets\.E2E_USER' 'secrets\.E2E_PASSWORD'; do
  ! grep -qE "$pat" "$wf" && ok "deploy.yml no longer references $pat" || ko "deploy.yml still references $pat"
done
! grep -qE 'vars\.' "$wf" && ok "deploy.yml reads no repository or environment variable" || ko "deploy.yml reads a variable"
mapfile -t secs < <(grep -oE 'secrets\.[A-Z0-9_]+' "$wf" | sort -u)
[ "${secs[*]}" = "secrets.DEPLOY_SSH_KEY secrets.SOPS_AGE_KEY" ] && ok "deploy.yml reads exactly two secrets: DEPLOY_SSH_KEY and SOPS_AGE_KEY" || ko "secrets read: ${secs[*]}"
grep -q 'scripts/deploy-conf.sh' "$wf" && ok "deploy.yml reads the settings with scripts/deploy-conf.sh" || ko "deploy-conf.sh not used"
grep -q 'StrictHostKeyChecking=yes' "$wf" && ! grep -q 'StrictHostKeyChecking=no' "$wf" && ok "host key checking stays strict" || ko "host key checking"
grep -q 'scripts/e2e-env.sh' "$wf" && ok "the e2e job reads the account from the sops file" || ko "e2e-env.sh not used"
# The settings step must come after the checkout of the commit to deploy (the file is read from that commit).
co="$(grep -n 'Check out the commit to deploy' "$wf" | head -1 | cut -d: -f1)"
st="$(grep -n 'scripts/deploy-conf.sh' "$wf" | head -1 | cut -d: -f1)"
[ -n "$co" ] && [ -n "$st" ] && [ "$st" -gt "$co" ] && ok "the settings are read after the checkout of the commit to deploy" || ko "order checkout/settings ($co/$st)"

finish
