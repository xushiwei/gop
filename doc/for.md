# The `for` Statement

XGo has a single looping keyword, `for`, with four forms: bare `for`, condition `for`, C-style `for`, and `for..in`. The first three are unchanged from Go. `for..in` is XGo-native syntax for iterating over a value, with an optional `if` filter clause. It is the recommended way to iterate in XGo, covering the cases Go handles with `for range`.

## Overview

Iterating "for each value" is the most common looping intent, and `for..in` is built around that intent directly: `for v in expr` binds `v` to each value in turn. When both a key/index and a value are needed, `for k, v in expr` binds both, in that order.

`for..in` operates over a wide range of iterable expressions: integers, arrays, slices, strings, maps, channels, range expressions (`start:end:step`), iterator functions, and any type that exposes an `XGo_Enum` method returning an iterator function.

XGo's class and object model benefits from a way for user-defined types to participate in `for..in` iteration without implementing `sort.Interface`-style multi-method protocols. The `XGo_Enum()` convention gives any type a single, minimal hook into `for..in`.

## Bare `for`

```go
for {
    // ...
}
```

The condition is omitted, producing an infinite loop. Use `break` or `return` to end it.

## Condition `for`

```go
sum := 0
i := 1
for i <= 100 {
    sum += i
    i++
}
echo sum // 5050
```

This form is equivalent to a `while` loop in other languages: the loop runs as long as the boolean condition evaluates to `true`. There are no parentheses around the condition, and the braces are always required.

## C-style `for`

```go
for i := 0; i < 10; i += 2 {
    if i == 6 {
        continue
    }
    echo i
    // 0
    // 2
    // 4
    // 8
}
```

The traditional three-clause form: init statement, condition, post statement, separated by `;`. It is safer than the condition `for` form because the counter update is part of the loop header and cannot be forgotten.

## `for..in`

### Syntax

`for..in` takes the form `for <identifiers> in <expr> { ... }`, with an optional `if <condition>` inserted between the iterated expression and the opening brace. Two identifier forms are supported:

```go
for v in expr { ... }     // binds the value
for k, v in expr { ... }  // binds the key/index and the value
```

`for..in` always declares new variables in the scope of the loop body. A variable that should not be used may be written as `_`, per normal XGo/Go rules.

`for..in` with zero identifiers (`for in expr`) is not valid; use `for expr {}`-style side-effecting constructs elsewhere if iteration is needed purely for side effects without touching the yielded values.

### Binding rule

When a single variable is bound, `for v in expr` binds the **value**. When two variables are bound, `for k, v in expr` binds the key/index first and the value second.

The rule is uniform across every iterable category: for any category that produces two components (a key/index and a value), `for v in expr` binds the value and discards the key/index; `for k, v in expr` binds both, key/index first. For categories that produce a single component, that component is treated as the value and bound directly by `for v in expr`; `for k, v in expr` is a compile error since there is no second component to bind.

### Optional `if` filter

`for..in` accepts an optional trailing `if` clause. When present, the loop body executes only for elements where the condition evaluates to `true`; elements that fail the condition are skipped without entering the body. The filter expression may reference the identifiers bound by the `in` clause.

```go
numbers := [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
for num in numbers if num%3 == 0 {
    echo num
    // 0
    // 3
    // 6
    // 9
}

for num in :10 if num%3 == 0 {
    echo num
    // 0
    // 3
    // 6
    // 9
}
```

## Iterable Expressions

`expr` in a `for..in` clause may be one of:

1. `int` — iterates the successive indices `0, 1, ..., n-1`. `n` is equivalent to the range expression `:n`, i.e. `int` is the special case of a range expression with `start` and `step` both left at their defaults.
2. `array`, `slice`, `string`, `map`, `chan` — the standard Go container/channel kinds. Iterating a `string` yields runes (with byte offsets as the key/index); iterating a `map` yields values (with keys as the key/index); iterating a `chan` yields the received values only, with no key/index.
3. A range expression, written `start:end`, `:end`, or `start:end:step` — an XGo-native numeric range, independent of any variable's type. `start` defaults to `0` when omitted; `step` defaults to `1` and must not be `0`. `start:end` produces the half-open sequence `start, start+1, ..., end-1`. `start:end:step` produces `start, start+step, start+2*step, ...` while the value remains less than `end` (for positive `step`) or greater than `end` (for negative `step`). `start` and `end` may be any expression of a common integer type; a range expression is not a value and may only appear directly as the `expr` in an `in` clause.
4. `func` matching one of the following iterator signatures:
   - `func(yield func() bool)`
   - `func(yield func(V) bool)`
   - `func(yield func(K, V) bool)`
5. Any type `T` (or `*T`) with a method `XGo_Enum() F`, where `F` is one of the three iterator function types listed above. `XGo_Enum` takes no parameters and returns exactly one value.

For case 5, the compiler rewrites the `in` clause to iterate over `expr.XGo_Enum()` instead of `expr` directly, then applies the rules for case 4.

Resolution order: the compiler first checks whether `expr`'s type is a built-in iterable kind (int, array, slice, string, map, chan, func) or a range expression. If not, it looks for an `XGo_Enum() F` method on the type or its pointer type. If neither applies, the expression is not iterable and compilation fails.

## Semantics by Category

The table below defines what `for v in expr` and `for k, v in expr` bind for each category.

| Category                        | Components produced        | `for v in expr` binds       | `for k, v in expr` binds     |
|----------------------------------|-----------------------------|-------------------------------|---------------------------------|
| `int`                             | index only                  | the index                      | invalid (only one value)        |
| `array` / `slice`                 | index, element                | the element                    | index, element                    |
| `string`                          | byte index, rune                | the rune                        | byte index, rune                   |
| `map`                              | key, value                    | the value                       | key, value                        |
| `chan`                             | value only                   | the value                       | invalid (only one value)        |
| Range expression (`start:end:step`) | successive stepped value | successive stepped value    | invalid (only one value)        |
| `func(yield func() bool)`         | (none)                       | invalid (no value)              | invalid (no value)                |
| `func(yield func(V) bool)`        | value only                   | the value                       | invalid (only one value)        |
| `func(yield func(K, V) bool)`     | key, value                    | the value (key discarded)     | key, value                        |
| `T` with `XGo_Enum() F`           | per `F` above                | per `F` above                  | per `F` above                     |

## Examples by Category

### Array, slice, string, map, chan

```go
for v in [1, 2, 3] {
    echo v               // 1, 2, 3
}

for i, v in [1, 2, 3] {
    echo i, v            // index, element
}

for k, v in {"a": 1, "b": 2} {
    echo k, v            // key, value
}

for v in {"one": 1, "two": 2} {
    echo v               // 1, 2
}

for c in "héllo" {
    echo c               // rune, not byte index
}

ch := ... // ch is a channel
for v in ch {
    echo v
}
```

`for k, v in ch` is a compile error: a channel produces only one value.

### int

`for i in 5` yields the successive indices `0, 1, 2, 3, 4`. It is equivalent to `for i in :5` (see below).

```go
for i in :5 {
    echo i
    // 0, 1, 2, 3, 4
}
```

`for k, v in 5` is a compile error: `int` produces only one value.

### Range expression

A range expression (`start:end:step`) is XGo-native. It produces successive values from `start` (default `0`) up to but excluding `end`, advancing by `step` (default `1`, must not be `0`; a negative `step` counts down).

```go
for v in 1:10 {
    echo v              // 1, 2, ..., 9
}

for v in :10 {
    echo v              // 0, 1, ..., 9 (start defaults to 0)
}

for v in :10:2 {
    echo v              // 0, 2, 4, 6, 8
}

for v in 10:0:-1 {
    echo v              // 10, 9, ..., 1
}
```

`for k, v in 1:10:2` is a compile error: a range expression produces only one value.

### func (iterator functions)

`for..in` binds the value component and discards the key, if any, when only one variable is given.

```go
for v in Seq {          // Seq: func(yield func(int) bool)
    echo v
}

for k, v in Seq2 {      // Seq2: func(yield func(string, int) bool)
    echo k, v
}

for v in Seq2 {         // key discarded; only the value is bound
    echo v
}
```

A zero-value iterator (`func(yield func() bool)`) has nothing to bind.

### Object with `XGo_Enum()`

Any type with an `XGo_Enum() F` method is iterable as if the expression were `expr.XGo_Enum()` itself — `for..in` applies the rules for whichever `func` signature `F` is.

```go
type Tree struct {
    // ...
}

func (t *Tree) XGo_Enum() func(func(int) bool) {
    return func(yield func(int) bool) {
        // in-order traversal, calling yield(v) for each element
    }
}

tree := new(Tree)
...
for v in tree {
    echo v
}
```

A two-value `XGo_Enum` follows the same two-value rules as above:

```go
type OrderedMap struct {
    // ...
}

func (m *OrderedMap) XGo_Enum() func(func(string, int) bool) {
    return func(yield func(string, int) bool) {
        // iterate in insertion order
    }
}

om := new(OrderedMap)
...
for k, v in om {
    echo k, v
}
```

### `for..in` with `if`

```go
numbers := [1, 3, 5, 7, 11, 13, 17]
sum := 0
for x in numbers if x > 5 {
    sum += x
}
echo sum // 48
```

The filter applies after the value (and key, if present) is bound, so it may reference either identifier:

```go
names := ["Sam", "Peter", "Alice"]
for i, name in names if i > 0 {
    echo i, name
    // 1 Peter
    // 2 Alice
}
```

## Restrictions

- `for k, v in expr` requires `expr`'s iteration protocol to produce two components (array, slice, string, map, `func(yield func(K, V) bool)`, or an `XGo_Enum()` returning one of these). Applying it to `int`, `chan`, a range expression, or `func(yield func(V) bool)` is a compile error.
- `for v in expr` requires at least one produced component. Applying it to `func(yield func() bool)` (or an `XGo_Enum()` returning that type) is a compile error.
- A range expression's `start`, if omitted, defaults to `0`; its `step`, if omitted, defaults to `1` and must not be `0` — a literal `0` step is a compile error, and a non-constant `0` step is a runtime error. A range expression may only appear directly as the `expr` of an `in` clause — it is not a first-class value and cannot be assigned, passed, or stored.
- `XGo_Enum` must take no parameters and return exactly one value, whose type is one of the three iterator signatures. A method named `XGo_Enum` with any other signature does not satisfy the protocol, and `expr` is treated as not iterable.
- Method-set rules are standard: an `XGo_Enum` defined on `*T` is visible through a value of type `T` only if that value is addressable, exactly as for any other pointer-receiver method.
- The `if` filter is evaluated once per iteration, after the loop variables for that iteration are bound and before the loop body runs.
