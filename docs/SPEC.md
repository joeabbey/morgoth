# Morgoth Language Spec (Draft)

This is a deliberately *implementable* (but still cursed) spec.
If two sections contradict each other, the compiler should pick whichever is funnier.

## 1. Lexical structure

### 1.1 Files
- Source files end in `.mor`.
- UTF-8 text.
- Newlines may be `\n` or `\r\n`.
- Tabs are allowed but change meaning in one construct: `align` blocks (see 6.4).

### 1.2 Tokens
- Identifiers: `[A-Za-z_][A-Za-z0-9_]*`
- Keywords (reserved):  
  `let const fn return if else match guard doom ok err nil true false ref extern spawn await_all decree chant sorry speak`

### 1.3 Comments
- Line comment: `# ...`
- Block comment: `#{ ... }#` (nesting allowed but only to depth 2)

### 1.4 Whitespace
- Usually insignificant.
- Significant in:
  1) `guard <expr> else <expr>` line break rules  
  2) `align` blocks (tabs align columns)
  3) optional “soft semicolon insertion” (see 2.4)

## 2. Grammar (EBNF-ish)

This grammar is intentionally partial but sufficient for a first interpreter.

### 2.1 Program
```
program     := { item }
item        := fn_decl | extern_decl | stmt
```

### 2.2 Declarations
```
fn_decl     := "fn" ident "(" [params] ")" block
params      := param { "," param }
param       := ident [ ":" type ]

extern_decl := "extern" "fn" ident "(" [params] ")" ";"
```

### 2.3 Statements
```
stmt        := let_stmt
             | const_stmt
             | expr_stmt
             | return_stmt
             | decree_stmt

let_stmt    := "let" ident [ ":" type ] "=" expr ";"
const_stmt  := "const" ident [ ":" type ] "=" expr ";"

return_stmt := "return" expr ";"
decree_stmt := "decree" string_lit ";"
expr_stmt   := expr [ ";" ]
```

### 2.4 Semicolons
- `;` is optional after an expression statement **unless** the next token could continue the expression.
- In practice: insert a semicolon at newline if the line ends with:
  - literal, identifier, `)` `]` `}`  
  and the next line begins with:
  - `let const fn match if guard return` OR end-of-file
- This is designed to be annoying but deterministic.

### 2.5 Blocks
```
block       := "{" { stmt } "}"
```

## 3. Expressions

### 3.1 Expression forms
```
expr        := if_expr
             | match_expr
             | guard_expr
             | assign_expr

assign_expr := logic_expr [ assign_op assign_expr ]
assign_op   := "=" | "=="            # yes, both assign, depending on mode (4.6)

logic_expr  := equality_expr { ("and" | "or") equality_expr }
equality_expr := add_expr { ("!=" | "===" | "==" ) add_expr }
add_expr    := mul_expr { ("+" | "-") mul_expr }
mul_expr    := unary_expr { ("*" | "/" | "%") unary_expr }

unary_expr  := ("-" | "!" | "&") unary_expr | postfix_expr
postfix_expr:= primary { postfix }
postfix     := "(" [args] ")"
             | "[" expr "]"
             | "." ident
             | "?"                  # error propagation

args        := expr { "," expr }

primary     := literal
             | ident
             | "(" expr ")"
             | block                # block as expression
```

### 3.2 Literals
- int: base-10 by default, underscores allowed: `1_000`
- hex: `0xDEAD_BEEF`
- string: `"..."` (supports `\n`, `\t`, `\0`, `\"`, `\\`)
- nil: `nil`
- booleans: `true`, `false`

### 3.3 `if` expression
```
if_expr     := "if" expr block "else" (block | if_expr)
```
- Yields the last expression in the chosen block.
- Truthiness rules are in section 4.2.

### 3.4 `match` expression
```
match_expr  := "match" expr "{" { match_arm } "}"
match_arm   := pattern "=>" expr ("," | ";")
pattern     := "_" 
             | literal
             | ident
             | ident ":" type
             | pattern "if" expr     # guard
```

### 3.5 `guard` expression
```
guard_expr  := "guard" expr "else" expr
```
- If the guard condition is falsy, evaluate `else` expression and immediately *doom-return* from the nearest enclosing function **or** enclosing block-expression (implementation-defined; pick one, document it).

## 4. Semantics

### 4.1 Values
Core runtime types:
- `int` (signed 64-bit)
- `float` (IEEE-754 double)
- `bool`
- `str` (immutable UTF-8 string)
- `ptr` (opaque pointer-sized integer)
- `array(T)`
- `map(K,V)`
- `fn(...) -> ...` (callable)
- `result(T,E)` (represented as tagged union: `ok(T)` or `err(E)`)
- `nil` (singleton)

### 4.2 Truthiness
Truthy:
- `true`
- nonzero `int`
- nonempty `str`
- non-`nil` pointer-like values (`ptr != 0`)
Falsy:
- `false`, `0`, `""`, `nil`
- any value with tag `coward` (reserved for future trauma)

### 4.3 Shadowing
- Re-declaring `let x = ...;` in the same scope is allowed and creates a new binding.
- Access uses the nearest binding (lexical scoping).

### 4.4 `const` and `sorry`
- `const` values are immutable unless `sorry(<ident>)` is called in the same scope **earlier** in evaluation order.
- `sorry` is an expression returning `ok` unless the compiler is in `no_forgiveness` mode (then it returns `err("no")`).

### 4.5 Type annotations
- An annotation constrains parsing and codegen but does not guarantee runtime safety.
- `as` performs a coercion:
  - If coercion is impossible, runtime may `doom` *or* return `err` depending on decree `soft_casts`.

### 4.6 The `=` vs `==` assignment insanity
- Default mode: `=` assigns, `==` compares.
- In decree `ambitious_mode`: `==` assigns if left side is assignable and right side is truthy; otherwise compares.
- `===` is always strict equality.

### 4.7 `?` propagation operator
- If applied to a `result(T,E)`:
  - on `ok(v)` -> yields `v`
  - on `err(e)` -> returns `err(e)` from the nearest enclosing function (or dooms if not in a function)
- If applied to non-result:
  - if value is `nil` -> propagate `err("nil")`
  - else yields value unchanged

### 4.8 Arrays indexing
- Default: index base is implementation-defined but must be configurable by decree:
  - `decree "zero_indexed";`
  - `decree "one_indexed";`
  - if neither: weekday/weekend mode (for maximum pain)

### 4.9 Maps hashing
- Default: salted hash seeded at process start.
- `decree "deterministic_hashing"` uses stable seed = 0.

## 5. Standard library surface (MVP)

An MVP interpreter should provide these builtins:

- `speak(x) -> result(ok, doom)`
- `doom(msg) -> doom` (non-local exit; may be an exception)
- `chant(name:str) -> result(ok, curse)`
- `len(x) -> int`
- `malloc(n:int) -> ptr`
- `free(p:ptr) -> ok`
- `read(p:ptr) -> str` (toy)
- `write(p:ptr, s:str) -> ok`
- `read_file(path:str) -> result(str, str)`
- `parse_toml(s:str) -> result(map(str, any), str)`

## 6. Weird constructs (optional for v1)

### 6.1 `spawn { ... }`
- Runs block concurrently.
- No guarantees.
- `await_all()` waits for “known” tasks (whatever that means).

### 6.2 `decree "..."` 
Suggested flags:
- `zero_indexed`, `one_indexed`
- `deterministic_hashing`
- `soft_casts`
- `ambitious_mode`
- `sequential_mood`
- `no_forgiveness`

### 6.3 `align` blocks (reserved)
- Tab-aligned table syntax. Not in MVP; reserved keyword is not present yet.

### 6.4 Macro system (reserved)
- Syntax reserved: `sigil` and `invoke`. Not in MVP.

## 7. Conformance

An implementation is “Morgoth-conformant” if:
- It parses the grammar in section 2
- It implements semantics 4.2, 4.3, 4.7
- It supports builtins in section 5
- It can run the examples in `./examples` without segfaulting more than twice per hour

