## Subscribe Action — <Provider> — [PR N: <name>]

Part of the Subscribe Action stack for `<provider>`. See [CONTRIBUTING_SUBSCRIBE_ACTION.md](../../CONTRIBUTING_SUBSCRIBE_ACTION.md).

- [ ] Scope limited to this stack rung (one interface / concern)
- [ ] Provider remains gated off (`Support.Subscribe` stays false) — except the Enable PR
- [ ] Any `SubscribeRequirements` flag set to `new(true)` has a code comment linking the provider docs
- [ ] Compile-time interface assertion added (if this PR adds an interface)
- [ ] Unit tests added/updated
- [ ] Manual sandbox verification (where applicable)
- [ ] Linked the relevant [SUBSCRIBE_REFERENCES.md](../../SUBSCRIBE_REFERENCES.md) section
