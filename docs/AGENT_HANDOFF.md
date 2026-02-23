# Mordor Implementation Handoff (for a coding agent)

This repository is a *spec + examples* handoff for implementing an MVP Mordor interpreter.

## Goals (MVP)

Implement a command-line tool:

- `mordor run <file.mor>`
- Optionally: `mordor repl`

The MVP must:
1) Parse Mordor source (`.mor`) according to `docs/SPEC.md` sections 1–3 (enough for examples).
2) Execute AST with runtime semantics:
   - truthiness (SPEC 4.2)
   - lexical scoping + shadowing (SPEC 4.3)
   - result type + `?` propagation (SPEC 4.7)
   - `guard ... else ...` non-local exit
3) Provide builtins (SPEC 5) sufficient for the examples.

Nice-to-have:
- decrees (`decree "..."`) affecting runtime behavior (indexing, hashing seed, ambitious_mode).
- `match` with basic patterns + guards.
- `spawn` as a stub that runs synchronously unless `sequential_mood` is absent.

## Recommended architecture

### Frontend
- Lexer -> tokens (track spans for error reporting)
- Parser -> AST (expression-centric)
- Desugar:
  - optional semicolons
  - `guard` into explicit branching + non-local exit node

### Runtime
- Value enum / tagged union
- Environment chain for scopes
- Execution stack frames:
  - functions
  - block-expressions (if you implement them as expressions)
- Errors:
  - distinguish `doom` (non-local exit / fatal) from `err(e)` (value-level result)
  - simplest: implement `doom` as raising an exception-like signal

### Decrees
- Store in a per-thread runtime config:
  - indexing_base: 0/1/weekday-weekend
  - hashing_seed: random/0
  - ambitious_mode: bool
  - soft_casts: bool
  - sequential_mood: bool
  - no_forgiveness: bool

## Minimal feature checklist (to pass examples)

- Literals: int, string, bool, nil
- `let`, `const`, assignment `=`
- `if ... else ...`
- function definitions + calls
- `match` with:
  - `_`, literal, `ident`, `ident: type` (type tag check), `pattern if expr`
- arrays + indexing
- maps + indexing
- `ok(x)`, `err(x)` constructors (either keywords or builtins)
- `?` operator
- `guard expr else expr`
- `speak`, `doom`, `len`

## Test approach

- Golden tests: run each file in `examples/` and compare stdout/stderr.
- Add fuzz tests for lexer/parser if time permits.

## Deliverables

1) Working `mordor` CLI.
2) `README` instructions for running examples.
3) A minimal standard library file or builtins implementation notes.

## Non-goals (for MVP)

- Real memory management (`malloc/free` can be fake)
- Concurrency correctness
- Optimizations
- A real installer
- A real package manager (Palantír is a lie)

