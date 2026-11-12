# secrets/

Encrypted secrets for the deployment profiles, with SOPS and age ([ADR-0009](../docs/decisions/0009-secrets-with-sops-and-age.md)).

- `<profile>.enc.env`: variable names in clear, every value encrypted. Committed.
- `<profile>.recipients`: the PUBLIC age keys allowed to decrypt that profile. Committed.
- Your PRIVATE key lives at `~/.config/sops/age/keys.txt` and is never committed.

```
scripts/secrets.sh keygen                  # once: create your key, print the public key to add to a .recipients file
scripts/secrets.sh init staging-test       # once per profile: encrypt the .env.example placeholders
scripts/secrets.sh edit staging-test       # change values (sops decrypts, you edit, sops re-encrypts)
scripts/secrets.sh decrypt staging-test    # only if you really want a plaintext .env.staging-test (mode 600, git-ignored)
scripts/secrets.sh rotate staging-test     # after a recipients change
```

**You do not decrypt to run the app.** `scripts/compose.sh <profile> ...` (and so `make up-local-dev`) decrypts `secrets/<profile>.enc.env` itself, at runtime, into a private RAM-backed temporary file that is removed when the command ends. It needs your private key (`~/.config/sops/age/keys.txt`) to be one of the file's recipients.

Only `*.enc.env`, `*.recipients` and this file belong here; `make security` fails on anything else.
