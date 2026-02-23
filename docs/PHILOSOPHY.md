# Mordor Philosophy

Mordor is not a language. It is a *boundary test*.

It’s what happens when you take:
- “the programmer should be in control”
and remove:
- “the programmer should be protected from themselves”

## The Pit of Despair (not the pit of success)

Most languages aim for a “pit of success”: do the right thing and you fall into good outcomes.

Mordor builds a “pit of despair”:
- do the wrong thing and you fall quickly
- do the right thing and you still fall, but with better error messages

## Explicitness is a trap

Mordor rejects the idea that explicit code is automatically correct.
You must prove correctness via:
- paranoia
- guard rails you wrote yourself
- testing
- and occasional bargaining

## Sharp edges are documentation

When Mordor cuts you:
- that’s the spec teaching you
- that’s the runtime showing you the boundary conditions
- that’s your future self leaving you a note

## “Undefined Behavior” is a social contract

The compiler makes no promise except:
- it will do *something*
- it will not apologize
- it will do the same thing again if you set the same decrees and the same moon phase

## Ergonomics are a luxury item

Mordor’s ergonomics are intentionally weird to force you to:
- slow down
- read closely
- stop assuming other languages’ rules apply

If your brain autocompletes “how this should work,” Mordor will punish you for it.

## Design goals (in plain terms)

- Make it possible to build an interpreter quickly.
- Make it possible to build a compiler later.
- Make it impossible to build complacency.

## The Mordor mindset

- Treat every operation as fallible.
- Treat every value as potentially cursed.
- Treat every convenience as a latent bug.

If you still want to use Mordor after reading this:
you’re probably the target audience.
