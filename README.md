# Operation Trails

Operation Trails help keeping track of an operation that crosses API boundaries and mediums

[![wercker status](https://app.wercker.com/status/9ce923f07a81073951d971a726515cac/m/master "wercker status")](https://app.wercker.com/project/byKey/9ce923f07a81073951d971a726515cac)


## Example Usage

Import with

`import "github.com/off-the-grid-inc/optrail"`

In its simplest usage, say you want to track an operation across several functions:

```go
func step1() {
    trail := optrail.Begin("step1", "info about this step, can be any type")

    // Do step1 work

    trail.Here("step1-b", "additional information computed in this function")

    step2(trail)
}

func step2(op *OpTrail) {
    trail.Here("step2", "info about this step, can be any type")

    // Do step2 work

    err := possiblyFailingCall()
    if err = trail.MaybeFail(err); err != nil {
        // Maybe show the error to the user, separately from the OpTrail lifecycle
        return
    }

    step3(trail)
}

func step3(trail *OpTrail) {
    trail.Here("step3", "info about this step, can be any type")

    // Do step3 work

    trail.Succeed()
}
```

### Forking

Forking means duplicating the trail into a new, independent one. It can be used when an operation becomes two.

*NOTE: trail inheritance is not implemented yet (critical)*

```go
func step() {
	op := Begin("test-op")
	op2 := op.Fork()

	op.Here("parent", "op1data")
	op2.Here("forked", "op2data")

	step2(op)
	step2(op2)
}
```


### Transmuting

Transmuting is used for temporarily following a trail within a different and new goroutine. While that happens, the original trail will belong in the new goroutine, only returning back to the original goroutine after it returns.

*NOTE: trail inheritance is not implemented yet (critical)*

```go
func step() {
	op := Begin("test-op")
    done := make(chan struct{})
    op.Transmute(func() {
        time.Sleep(2*time.Second)
        op.Here("in-new-goroutine", "info from goroutine")
        done <- struct{}{}
    })

    <-done
    op.Here("in-old-goroutine", "info from old goroutine")
    op.Succeed()
}
```

## TODO

- Child and spawned OpTrails don't inherit the parent context at the moment
