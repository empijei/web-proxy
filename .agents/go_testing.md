# Using tst for Go Tests

You MUST use `github.com/empijei/tst` to write lean tests.

## Key Helpers

- `v := tst.Do(f())(t)`: Unwraps `(V, error)`; `t.Fatalf` on error.
- `v := tst.DoB(f())(t)`: Unwraps `(V, bool)`; `t.Fatalf` on false.
- `v1, v2 := tst.Do2(f())(t)`: Unwraps `(V1, V2, error)`; `t.Fatalf` on error.
- `tst.No(err, t)`: `t.Fatalf` if `err != nil`.
- `tst.Is(want, got, t)`: `t.Errorf` on mismatch (uses `go-cmp`, accepts cmp.Option as vararg).
- `tst.Be(ok, t)`: `t.Fatalf` on false.
- `tst.Err("sub", err, t)`: Asserts `err` contains substring.
- `tst.Ko(t)`: `t.Fatalf` if `t.Failed()`. Stop cascading errors.
- `ctx := tst.Go(t)`: Calls `t.Parallel()` and returns `t.Context()`.

## Pattern

Prefer `tst.Do` over `if err != nil` for setup. Use `tst.Is` for all value comparisons.

There are no negations helpers such as "NotIs" or "BeNot", all the available functions are listed in this file.
