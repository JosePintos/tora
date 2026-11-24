# context-integration

### Decision

Bounded contexts talk to each other only through ports owned by the context that needs the information. The calling context defines a small interface in its own `application/ports`, in its own words. An adapter in its own `infrastructure/` implements it by calling the other context's application service, and translates that context's errors into its own.

The first case: the card context needs to know whether a user owns a deck. It defines `DeckAccess.EnsureOwnsDeck`. The `card/infrastructure/deckaccess` adapter implements it with `DeckService.GetOwned` and turns `deckservices.ErrDeckNotFound` into the card context's `ports.ErrDeckNotFound`.

Resources owned by another user are reported as `NOT_FOUND`, exactly like missing ones.

### Status

Accepted

### Reason

- The card service depends on one method it defines, not on the deck context's whole service. That keeps the contexts loosely coupled, which matters if one is ever extracted from the modular monolith (see `monolith.md`).
- The deck context stays the only owner of the ownership rule (`Deck.IsOwnedBy`). Card doesn't copy it or query the `decks` table itself.
- Card tests mock one small interface.
- The adapter is the anti-corruption layer: changes to deck errors or types stop there.
- Answering `NOT_FOUND` for other users' decks means deck IDs can't be probed, the same reasoning as the login enumeration fix.

### Alternatives

- **Inject `*deckservices.DeckService` into `CardService` directly.** Less code, but the card application layer would depend on another context's concrete service and every method on it.
- **Check ownership in SQL** (insert or select joined with `decks` and filtered by `owner_id`). One query instead of two, but the authorization rule moves into SQL, is duplicated per query, and can't be unit-tested.
- **Answer `FORBIDDEN`.** More precise, but it confirms to the caller that the deck exists.

### Consequences

- Creating a card costs one extra query, the ownership check. That's acceptable at this scale. If it shows up in profiling, a combined SQL query can replace the adapter's implementation without changing the port.
- The check and the insert aren't one transaction. A deck deleted in between makes the insert fail on the foreign key, which reports as an internal error. That's acceptable while decks can't be deleted.
