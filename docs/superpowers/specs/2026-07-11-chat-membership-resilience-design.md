# Chat Membership Resilience Design

## Goal

Fix chat participant pagination requests and make repeated membership actions
idempotent, while leaving reply/update polling unchanged.

## Scope

- `chat participants --offset N` must be accepted by the strict JSON RPC
  contract and forwarded to Telegram for channels.
- Basic-group participant results must apply the requested offset locally,
  because `messages.getFullChat` returns the participant collection directly.
- `USER_ALREADY_PARTICIPANT` from join, subscribe, or invite operations must
  be treated as an already-achieved membership state.
- Polling intervals and polling behavior are explicitly out of scope.

## Design

### Participant pagination

Add `Offset int` to `types.GetParticipantsParams`. Normalize negative offsets
to zero. For channels, set `Offset` on `tg.ChannelsGetParticipantsRequest`.
For basic groups, preserve the total participant count, skip the first Offset
items, and return at most the requested Limit items. The existing CLI
`HasOffset` surface remains enabled.

The same parameter type is reused by `GetAdmins`; its existing behavior stays
compatible, with the offset available to the underlying participant query.

### Idempotent membership actions

Use gotd's typed `tg.IsUserAlreadyParticipant` predicate before wrapping errors.

- `Join` returns a successful nil update when the account is already in the
  invite target; `JoinChat` therefore returns its normal successful result.
- `Subscribe` does the same for an already-joined channel;
  `SubscribeChannel` returns its normal successful result.
- `Invite` ignores this error for an individual member and continues inviting
  the remaining members. Non-idempotent errors still fail the operation.

No global IPC error-classification change is needed: these cases become
successful domain results rather than errors reaching the classifier.

## Testing

- Verify channel participant requests receive the requested offset and limit.
- Verify basic-group participant pagination skips the requested prefix and
  preserves total count.
- Verify join and subscribe return success for a typed
  `USER_ALREADY_PARTICIPANT` error.
- Verify invite continues after an already-participant member and still
  reports failure for unrelated errors.
- Run the affected package tests and the full Go test suite.
