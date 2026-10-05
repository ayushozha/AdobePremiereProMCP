# Sequence-list contract

Both CEP host entrypoints now use the native `SequenceCollection.numSequences` count and emit the public snake_case sequence fields: identity, dimensions, timebase, track counts, and active-sequence state. Unreadable metadata returns an error instead of fabricated zero counts.

The Go consumer accepts complete older camelCase responses and legacy success/data wrappers, while preserving snake_case output. It rejects incomplete inventories, conflicting aliases, duplicate identities, inconsistent counts, and an active sequence that cannot be verified. Reload older panels that return only names and IDs.

Verification: executable host fixtures cover both core and full hosts; Go consumer fixtures cover current and older response shapes, empty projects, and invalid inventories. These automated fixtures do not prove native compatibility with every Premiere version.
