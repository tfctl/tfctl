# Release signing with Cosign

This project uses Sigstore's `cosign` for signing release artifacts.

Why sign?

- Verifies integrity: users can confirm the downloaded binary matches what was released.
- Verifies authenticity: signatures show the artifact was released by a trusted party.

How releases are signed

- CI uses keyless Cosign signing (OIDC) to sign artifacts produced by GoReleaser.
- Detached signatures (`.sig`) are uploaded to the GitHub Release alongside the artifacts.
- Optionally, if you provide a `COSIGN_PUB` repository secret containing your public key, the workflow will attach `cosign.pub` to the release.

Local verification

- Verify a keyless-signed artifact:

```
cosign verify-blob --keyless path/to/artifact
```

- Verify a key-based signature when you have the public key:

```
cosign verify-blob --key cosign.pub path/to/artifact
```

Publishing the public key

- If you maintain a long-lived signing key and want users to verify releases with it, set the `COSIGN_PUB` secret (ASCII armored public key) in the repository and the workflow will publish it to each release.

Security notes

- Keyless signing (OIDC) is recommended for CI since it avoids long-lived private keys in secrets.
- If you use a private key for signing, store and rotate it securely and keep the private key out of the repo.
