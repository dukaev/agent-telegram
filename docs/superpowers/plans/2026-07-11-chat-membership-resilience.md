# Chat Membership Resilience Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task with review checkpoints.

**Goal:** Make participant-list offsets valid across the strict RPC contract, treat already-achieved membership actions as success, and release the fix as v0.7.17.

**Architecture:** Keep pagination in the chat domain layer: channel offsets go directly to Telegram, while legacy-group results are sliced locally. Normalize `USER_ALREADY_PARTICIPANT` before generic IPC error classification so join, subscribe, and invite remain idempotent without changing global error semantics.

**Tech Stack:** Go, gotd Telegram TL client, Cobra CLI, strict JSON RPC, npm package metadata.

## Global Constraints

- Reply/update polling is out of scope and must not be changed.
- Unknown JSON fields remain rejected by `internal/strictjson`.
- Non-idempotent Telegram errors retain their existing behavior.
- The patch release version is `0.7.17`.

---

### Task 1: Support participant offsets

**Files:**
- Modify: `telegram/types/types_chat_members.go`
- Modify: `telegram/chat/participants.go`
- Test: `telegram/chat/chat_test.go`

**Interfaces:**
- Consumes: `types.GetParticipantsParams`, `types.GetAdminsParams`, `types.GetBannedParams` and existing `HasOffset` CLI parameters.
- Produces: accepted `offset` JSON for all three participant-list methods; channel requests with `Offset`; locally paginated legacy-group results.

- [ ] **Step 1: Add offset fields to the public participant contracts**

Add `Offset int \`json:"offset,omitempty"\`` to `GetParticipantsParams`, `GetAdminsParams`, and `GetBannedParams`. Keep the field names and JSON tags identical so the existing `types.GetParticipantsParams(params)` conversion for admins remains valid.

- [ ] **Step 2: Implement channel and legacy-group pagination**

Normalize negative offsets to zero. Set `Offset: params.Offset` on `tg.ChannelsGetParticipantsRequest` in `GetParticipants` and `GetBanned`. In the basic-group branch of `GetParticipants`, set `Count` from the full participant collection, then slice from `params.Offset` and cap the returned page at `limit`; an offset beyond the collection returns an empty page.

- [ ] **Step 3: Add deterministic fake-API tests**

Extend `telegram/chat/chat_test.go` with a fake `client.ParentClient`. Assert a channel request receives `Offset: 2` and `Limit: 2`, and assert a legacy-group request returns only the page after offset while preserving the total count. Also assert JSON decoding of `offset` into each public parameter type succeeds through the existing strict handler path if a focused contract test is practical.

- [ ] **Step 4: Run the focused tests**

Run: `go test ./telegram/chat ./telegram/types -count=1`

Expected: PASS.

### Task 2: Make membership actions idempotent

**Files:**
- Modify: `telegram/chat/join.go`
- Modify: `telegram/chat/members.go`
- Test: `telegram/chat/chat_test.go`

**Interfaces:**
- Consumes: gotd's `tg.IsUserAlreadyParticipant(error)` predicate.
- Produces: successful `JoinChat`/`SubscribeChannel` results for already-joined accounts; `Invite` continues past already-participating members.

- [ ] **Step 1: Add typed-error regression cases**

Use `tgerr.New(400, tg.ErrUserAlreadyParticipant)` in fake API invokers. Cover invite-link join and channel subscribe as successful operations, and cover a two-member invite where the first member returns `USER_ALREADY_PARTICIPANT` and the second succeeds.

- [ ] **Step 2: Normalize join and subscribe errors before wrapping**

In `Join` and `Subscribe`, when the Telegram call returns an error, check `tg.IsUserAlreadyParticipant(err)` before constructing the current contextual error. Return `nil, nil` for that case; preserve the existing wrapped error for every other case.

- [ ] **Step 3: Ignore only the idempotent invite error**

In each channel/chat invite call, continue to the next member when `tg.IsUserAlreadyParticipant(err)` is true. Return the existing wrapped error for all other failures. A list containing only existing members still returns `InviteResult{Success: true}`.

- [ ] **Step 4: Run focused membership tests**

Run: `go test ./telegram/chat -run 'Test.*(Join|Subscribe|Invite|Participant)' -count=1`

Expected: PASS.

### Task 3: Bump release metadata and verify the repository

**Files:**
- Modify: `package.json`
- Modify: `package-lock.json`

**Interfaces:**
- Consumes: current release version `0.7.16`.
- Produces: package version `0.7.17` in both npm metadata files.

- [ ] **Step 1: Bump package metadata**

Change the root package version from `0.7.16` to `0.7.17` and update the matching root `version` entry in `package-lock.json` without changing dependency versions.

- [ ] **Step 2: Format and run the full test suite**

Run: `gofmt -w telegram/chat/participants.go telegram/chat/join.go telegram/chat/chat_test.go telegram/types/types_chat_members.go telegram/chat/members.go`

Then run: `go test ./... -count=1`

Expected: all Go packages pass.

- [ ] **Step 3: Review the final diff**

Run: `git diff --check` and `git status --short`.

Expected: no whitespace errors; only the planned Go files, package metadata, and plan/spec documentation are changed.

- [ ] **Step 4: Commit the implementation**

```bash
git add telegram/types/types_chat_members.go telegram/chat/participants.go telegram/chat/join.go telegram/chat/members.go telegram/chat/chat_test.go package.json package-lock.json docs/superpowers/plans/2026-07-11-chat-membership-resilience.md
git commit -m "fix: make chat membership actions idempotent"
```

- [ ] **Step 5: Push and open the pull request**

```bash
git push -u origin codex/chat-membership-resilience
gh pr create --draft --title "fix: make chat membership actions idempotent" --body "Fix strict participant offset handling, make already-joined membership actions idempotent, and bump the package release to v0.7.17. Reply/update polling remains unchanged."
```

Expected: the branch is pushed and a draft PR URL is returned.
