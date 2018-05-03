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
    if err = trail.MaybeFail(err) != nil {
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


## TODO

- Child and spawned OpTrails don't inherit the parent context at the moment
