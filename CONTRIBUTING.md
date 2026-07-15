# Contributing

Contributions are welcome through focused pull requests that preserve strict behavior and clearly document security tradeoffs.

## Development setup

```bash
git clone https://github.com/MeViksry/quuid.git
cd quuid
go test ./...
go test -race ./...
go vet ./...
```

Run the complete local suite:

```bash
make check
```

## Pull-request requirements

A pull request should:

- remain limited to one coherent change;
- include regression tests for defects;
- include success, malformed-input, and boundary tests for new APIs;
- update README and package documentation for exported behavior;
- add fuzz seeds when parsing or decoding changes;
- preserve strict defaults;
- avoid hidden global configuration;
- avoid custom cryptographic primitives;
- explain compatibility impact and migration steps;
- pass formatting, unit tests, race tests, vet, build, and legacy-hash checks.

## Parser changes

Parser changes require particular care. Document every newly accepted format. Compatibility formats belong in explicitly named loose or migration APIs, not the strict parser.

Run:

```bash
go test -fuzz='^FuzzParseUUID$' -fuzztime=30s .
go test -fuzz='^FuzzParseUUIDLoose$' -fuzztime=30s .
go test -fuzz='^FuzzParseID$' -fuzztime=30s .
go test -fuzz='^FuzzParseStrongID$' -fuzztime=30s .
go test -fuzz='^FuzzParseSortableID$' -fuzztime=30s .
```

## Cryptographic changes

Do not introduce a new cryptographic primitive merely to avoid a dependency or produce a novel format. Use reviewed standard-library primitives and document domain separation, output size, and threat assumptions.

Changes involving randomness must test:

- nil readers;
- short reads;
- all-zero test sources;
- version and variant bits;
- concurrent use where applicable.

## Database changes

Database behavior must remain explicit per type. Do not add a package-global flag that changes whether UUID values are written as text or binary.

New SQL wrappers should test:

- `driver.Value` concrete type;
- exact binary length;
- text and binary scanning;
- SQL `NULL` behavior;
- empty string and empty byte behavior for nullable types;
- receiver state after malformed input.

## Commit style

Use imperative, scoped subjects:

```text
Add binary UUID SQL wrapper
Reject whitespace in strict UUID parser
Test RFC UUIDv6 example
Document monotonic UUIDv7 guarantees
```

## Compatibility

Before v1.0, a breaking refinement requires a clear changelog entry. After v1.0, exported API changes follow semantic versioning.
