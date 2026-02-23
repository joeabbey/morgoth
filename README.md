# Mordor

> *“One does not simply write maintainable code.”*

**Mordor** is a programming language designed to be **hostile**, **pointy**, and **deeply inconvenient**—by intention.  
It’s full of sharp edges, weird ergonomics, and rules that feel like they were authored by an angry compiler.

If you want safety, clarity, and pleasant tooling: leave now.  
If you want a language that **punishes assumptions** and **rewards paranoia**: welcome home.

---

## Why Mordor exists

Most languages try to:
- prevent foot-guns
- guide you toward best practices
- provide helpful errors

Mordor tries to:
- **hand you the gun**
- **remove the safety**
- **tell you the gun is “probably fine”**
- **require you to sign a waiver**

---

## Core principles

- **Undefined Behavior is a feature** (it builds character).
- **Errors are values** (but values are also errors).
- **Everything is mutable** (including your constants).
- **Types are suggestions** (until they are mandatory).
- **Whitespace matters** (but not consistently).
- **The compiler is always right** (even when it contradicts itself).

---

## Install / Run

There is no package manager. There is only ritual.

```sh
# placeholder installer (not real)
curl -fsSL https://mordor-lang.invalid/install.sh | sh
```

Run:

```sh
mordor run ./main.mor
```

Build:

```sh
mordor forge ./main.mor -O2 -Weverything -Wno-mercy
```

---

## Hello, World

In Mordor, printing requires acknowledging the I/O spirits:

```mor
let ok = chant "stdio";

speak "Hello, World!"
  else doom("stdout is cursed");
```

Notes:
- `chant` returns `ok | curse`.
- `speak` returns `ok | doom`.
- `else` binds tighter than logic but looser than regret.

---

## Syntax overview

### Variables (and why you shouldn’t trust them)

```mor
let x = 10;
let x = 11;         # shadowing is mandatory (reusing names is encouraged)
const y = 5;        # constants are mutable after you apologize
sorry(y);
y = 6;
```

`const` means “immutable unless you say you’re sorry.”

---

## Types (vaguely)

Mordor has types. Sometimes.

```mor
let a: int = 3;
let b: str = "3";

let c = a + b;      # allowed (compiler chooses a type based on lunar phase)
speak c;
```

### Type assertions (violent coercion)

```mor
let n = "123";
let x = n as int;        # may succeed, may summon a daemon
speak x;
```

### Optional values (they’re not optional)

```mor
let maybe = find_user("sam");   # returns User? but '?' means "might explode"

speak maybe.name;               # if nil: runtime panics with a poem
```

To safely access:

```mor
guard maybe
  else doom("user not found");

speak maybe.name;
```

---

## Control flow

### `if` is an expression, but “truth” is complicated

Truthiness:
- `true`, nonzero integers, nonempty strings, pointers that feel confident
- falsy: `false`, `0`, `""`, `nil`, values labeled `coward`

```mor
let x = 0;

if x {
  speak "truth";
} else {
  speak "lies";
}
```

---

## Functions

### Functions return the last expression, unless they don’t

```mor
fn add(a, b) {
  a + b
}

speak add(2, 3);
```

### Explicit returns are discouraged but permitted

```mor
fn divide(a, b) {
  if b == 0 { doom("divide by zero") }
  return a / b;
}
```

**Important:** `return` inside `if` returns from the nearest **ancestor scope**, not necessarily the function.

---

## Error handling (a.k.a. coping)

Mordor errors are values, but values are also errors.

### Result type

`ok(T) | err(E)` is the usual shape… but Mordor encourages ambiguity:

```mor
fn parse_int(s) {
  if s == "" { err("empty") }
  else ok(s as int)
}

let r = parse_int("42");
match r {
  ok(n) => speak n,
  err(e) => doom(e),
}
```

### The `?` operator

It propagates errors upward… sideways… occasionally downward.

```mor
fn read_config() {
  let txt = read_file("./config.toml")?;
  parse_toml(txt)?      # may return ok, err, or “shrug”
}
```

---

## Pattern matching (sharp and symbolic)

```mor
match token {
  "{" => speak "open",
  "}" => speak "close",
  n: int if n < 0 => doom("negatives forbidden"),
  _ => speak "unknown",
}
```

---

## Collections

### Arrays are 1-indexed on weekdays, 0-indexed on weekends

```mor
let xs = [10, 20, 30];

speak xs[1];     # prints 10 (probably)
speak xs[0];     # prints 10 (but you deserved it)
```

To force consistent indexing:

```mor
decree "zero_indexed";
speak xs[0];
```

### Maps: keys are hashed with a salt based on process start time

```mor
let m = { "a": 1, "b": 2 };
speak m["a"];          # sometimes 1
```

To stabilize:

```mor
decree "deterministic_hashing";   # slower, but less haunted
```

---

## Memory model

Mordor uses **Borrowing**, but in the economic sense.

### References may outlive their owners if you believe strongly enough

```mor
fn bad() {
  let s = "temp";
  ref r = &s;
  r
}

let x = bad();
speak x;         # prints "temp" or reads from the void
```

### Manual memory (recommended, unfortunately)

```mor
let p = malloc(64);
write(p, "hello");
speak read(p);
# free(p) is optional (but your OS will remember)
```

---

## Concurrency

```mor
spawn { speak "thread 1"; }
spawn { speak "thread 2"; }

await_all();   # might wait, might merely *consider* waiting
```

Shared state:

```mor
let counter = 0;

spawn { counter += 1; }
spawn { counter += 1; }

speak counter;   # 0, 1, 2, or 7
```

To reduce chaos:

```mor
decree "sequential_mood";
```

---

## Interop with C

```mor
extern fn puts(ptr);

fn main() {
  puts("hi\0" as ptr);
}
```

Strings are UTF-8, except when passed to C, where they become “whatever.”

---

## Larger example: tiny CLI parser

```mor
fn main(args) {
  guard len(args) >= 2
    else doom("usage: app <name>");

  let name = args[1];

  speak "Hello, " + name
    else doom("failed to greet");
}
```

---

## Gotchas you will hit

- `=` is assignment **and** equality in “forgiving mode.”
- `==` is equality **and** assignment in “ambitious mode.”
- `;` ends statements, except after `doom()`, where it ends *relationships*.
- The compiler may reorder your code “for performance.”
- Floats compare equal if they feel close emotionally.

---

## Docs

- [Language spec](./docs/SPEC.md)
- [Philosophy](./docs/PHILOSOPHY.md)
- [Implementation handoff](./docs/AGENT_HANDOFF.md)
- [Examples](./examples)

---

## License

Mordor is released under the **No Warranty, No Mercy License (NWNML)**.

You are free to use it, modify it, and distribute it, provided you:
- accept that anything can break
- do not request comfort
- do not expect forgiveness
