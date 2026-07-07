# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

While the module is pre-1.0 (`v0.x`), the public API is not yet stable: minor
version bumps may contain breaking changes. Breaking changes are always called
out under **Changed** / **Removed** so consumers pinning a version can review
them before upgrading.

## [Unreleased]

## [0.1.0] - 2026-07-08

First tagged release. Prior to this tag, consumers could only depend on Go
pseudo-versions of `main`; the breaking changes listed under **Changed** landed
during that pre-tag development and are recorded here so anyone upgrading from
an earlier pseudo-version is aware of them.

### Added

- GitHub Actions CI: `go build`, `go vet`, a `gofmt` gate, `go test -race` with
  coverage, plus `govulncheck` and `golangci-lint` (the latter two run
  non-blocking for now).
- Dependabot configuration for GitHub Actions and Go modules.
- `utils.ReadNotifyBody` — a shared, size-capped reader for inbound webhook
  bodies, reused by every notify handler.
- `offiaccount` access_token self-heal: a `40001`/`40014`/`42001`/`42007`
  response transparently invalidates the cached token and retries once.
- Systemic fail-fast input validation across `mini-store`, `xiaowei`,
  `mini-game`, `channels`, `oplatform`, and `isv`.

### Changed

- **Breaking — `merchant/types`:** all monetary fields widened from `int` to
  `int64` to avoid overflow on 32-bit platforms and to match WeChat Pay's
  amount range. Callers reading or assigning these fields must update their
  types.
- **Breaking — `TokenSource` / access_token plumbing hoisted to `utils`:**
  `TokenSource`, `FetchAccessToken`, and the shared `TokenCache` now live in
  `utils`. Each product package exposes `TokenSource = utils.TokenSource` as a
  type alias, so a single implementation (e.g. an Open-Platform authorizer)
  can serve every WeChat product line. Packages that previously declared their
  own token-source type are affected.
- **Breaking — `offiaccount.GetMaterial` refactored** as part of the same
  consolidation; review call sites if you used the previous signature.
- Relicensed under Apache-2.0.

### Fixed

- **`offiaccount.GetUserInfo` no longer swallows business errors.** `UserInfo`
  now embeds `Resp`, so a WeChat `errcode` (e.g. `40003`) surfaces as an error
  instead of returning a zero-valued `UserInfo` with a nil error — which had
  let an empty OpenID propagate into caller-side authorization logic.

### Security

- **Inbound webhook body size cap (audit H-1).** The `offiaccount`,
  `oplatform`, and `isv` notify handlers previously read the request body with
  an unbounded `io.ReadAll`, allowing an unauthenticated caller to trigger an
  out-of-memory condition. Bodies are now capped (default 1 MiB via
  `utils.ReadNotifyBody`) *before* any XML/JSON parsing.
- **WeChat Pay sensitive-field encryption corrected to SHA-1 OAEP (audit
  H-2).** `merchant/developed` was encrypting sensitive fields with
  `RSA-OAEP-SHA256`; WeChat Pay v3 mandates `RSA/ECB/OAEPWithSHA-1AndMGF1Padding`
  (SHA-1). Ciphertext produced by the old code could not be decrypted by
  WeChat. The fix restores SHA-1 and adds a reverse-locking test asserting
  SHA-256 decryption fails, to prevent a self-consistent-but-wrong regression.
- URL credential redaction (`utils.RedactURL`) and a default 10 MiB response
  body cap are applied across the shared HTTP client.

[Unreleased]: https://github.com/godrealms/go-wechat-sdk/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/godrealms/go-wechat-sdk/releases/tag/v0.1.0
