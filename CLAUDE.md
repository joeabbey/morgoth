# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Repo Is

Mordor is a **spec-only** repository for a deliberately hostile programming language. There is no implementation yet — the task is to build an MVP interpreter from scratch. The spec, philosophy, and example programs are all provided; your job is to write the interpreter.

## Key Documents

- `docs/SPEC.md` — Authoritative language specification (sections 1–5 are MVP scope)
- `docs/AGENT_HANDOFF.md` — Implementation checklist, recommended architecture, and non-goals
- `docs/PHILOSOPHY.md` — Design intent (sharp edges are intentional, not bugs)
- `examples/*.mor` — Six example programs that serve as the acceptance test suite

## Running Examples (Once Interpreter Exists)

```bash
mordor run examples/hello.mor
# or run all:
./scripts/run_examples.sh
```

Build output goes to `/build` or `/dist` (per `.gitignore`).

## MVP Requirements

The interpreter must provide a `mordor run <file.mor>` CLI that can execute all six example files. Required features (derived from the examples):

1. **Literals**: int, string, bool, nil
2. **Bindings**: `let`, `const`, assignment `=`, shadowing
3. **`const` + `sorry`**: `const` is immutable unless `sorry(ident)` is called first in the same scope
4. **Control flow**: `if`/`else` (as expressions), `match` with `_`/literal/ident/typed/guarded patterns
5. **Functions**: `fn` declarations, calls, implicit return (last expression)
6. **`guard expr else expr`**: falsy guard triggers non-local doom-return from enclosing function
7. **`?` operator**: propagates `err(e)` upward from `result` values; on `nil`, propagates `err("nil")`
8. **`ok(x)` / `err(x)`**: result type constructors
9. **Collections**: arrays (with decree-controlled indexing base), maps (with decree-controlled hashing)
10. **`decree "..."`**: runtime config flags — `zero_indexed`, `one_indexed`, `deterministic_hashing`, `soft_casts`, `ambitious_mode`, `sequential_mood`, `no_forgiveness`
11. **Builtins**: `speak`, `doom`, `chant`, `len`, `sorry`

## Recommended Architecture

Per `docs/AGENT_HANDOFF.md`:

**Frontend**: Lexer (with span tracking) → Parser (expression-centric AST) → Desugarer (semicolon insertion, guard → branching + non-local exit)

**Runtime**: Value enum/tagged union → Environment chain (lexical scoping) → Stack frames (functions + block-expressions) → Two error kinds: `doom` (non-local/fatal, like an exception) vs `err(e)` (value-level result)

**Decrees**: Per-thread runtime config struct with fields: `indexing_base`, `hashing_seed`, `ambitious_mode`, `soft_casts`, `sequential_mood`, `no_forgiveness`

## Semantic Gotchas to Get Right

- `=` is assignment; `==` is equality — **unless** `ambitious_mode` is decreed, where `==` assigns if the LHS is assignable and RHS is truthy. `===` is always strict equality.
- Semicolons are optional with deterministic newline-insertion rules (SPEC 2.4): insert `;` at newline if line ends with literal/ident/`)` `]` `}` and next line starts with a keyword or EOF.
- `speak` returns `result(ok, doom)` — the `else` clause after `speak` handles the error case.
- Array indexing defaults to weekday=1-indexed / weekend=0-indexed if no decree is set.
- Map hashing is salted by process start time unless `deterministic_hashing` is decreed.
- `chant` is an import/init builtin returning `result(ok, curse)` — for MVP, just return `ok`.

## Test Approach

- **Golden tests**: run each `examples/*.mor` file and compare stdout/stderr against expected output.
- Optional: fuzz tests for lexer/parser.

## Non-Goals (MVP)

- Real memory management (`malloc`/`free` can be stubs)
- Real concurrency (`spawn` can run synchronously)
- Optimizations
- A real installer or package manager
