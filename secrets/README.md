# secrets/

Encrypted secrets for the deployment profiles, with SOPS and age ([ADR-0009](../docs/decisions/0009-secrets-with-sops-and-age.md)).

- `<profile>.enc.env`: variable names in clear, every value encrypted. Committed.
- `<profile>.recipients`: the PUBLIC age keys allowed to decrypt that profile. Committed.
- Your PRIVATE key lives at `~/.config/sops/age/keys.txt` and is never committed.
- Profiles: `local-dev`, `staging-test`, `production-live`. `local-dev` secrets can be created with `scripts/secrets.sh init local-dev`. For `production-live`, the recipients are the owner plus the deployment host key only (ADR-0009, Decision 2).

```
scripts/secrets.sh keygen                  # once: create your key, print the public key to add to a .recipients file
scripts/secrets.sh init staging-test       # once per profile: encrypt .env.example; the database password, its copy in DATABASE_URL and CURSOR_KEY are generated, APP_ENV is set to the profile
scripts/secrets.sh edit staging-test       # change values (sops decrypts, you edit, sops re-encrypts)
scripts/secrets.sh decrypt staging-test    # only if you really want a plaintext .env.staging-test (mode 600, git-ignored); refuses to overwrite an existing file unless you add --force
scripts/secrets.sh rotate staging-test     # after a recipients change
scripts/secrets.sh check                   # verify nothing under secrets/ is plaintext (scripts/check-secrets-encrypted.sh; also part of make security)
```

**You do not decrypt to run the app.** `scripts/compose.sh <profile> ...` (and so `make up-local-dev`) decrypts `secrets/<profile>.enc.env` itself, at runtime, into a private RAM-backed temporary file that is removed when the command ends. It needs your private key (`~/.config/sops/age/keys.txt`) to be one of the file's recipients.

## New teammate

A new teammate cannot decrypt anything until an existing recipient adds their public key and rotates (ADR-0009, Consequences). Steps:

1. The teammate runs `scripts/secrets.sh keygen` and keeps the private key local.
2. They send only the PUBLIC key to the owner (issue #46 tracks collaborator keys).
3. The owner adds it to `secrets/<profile>.recipients` and runs `scripts/secrets.sh rotate <profile>`, then opens the PR.

Removing a person works the same way in reverse, and the secrets they could read must also be changed (ADR-0009, Decision 7).

Only `*.enc.env`, `*.recipients` and this file belong here; `make security` fails on anything else.
