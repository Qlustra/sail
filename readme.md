# Sail

Sail lets Go framework authors use the model their framework already needs as
the grammar for an application’s struct declarations. Write the grammar once,
then let Sail validate application declarations and fill that same model.

## What it does

- Compiles a tagged **Legend** into a reusable **Navigator**.
- Interprets an application’s **Chart** type and returns a fresh populated Legend.
- Visits populated occurrences, selects an immediate step, or finds complete routes.
- Records where each occurrence came from so a framework can resolve its live Chart field separately.

## Why it exists

A framework shouldn’t have to maintain a schema, a reflection parser, and a
destination model that all describe the same structure. With Sail, the model
defines what’s allowed; the framework still decides what the model means and
what to do with it. Sail does not parse CLI arguments or dispatch operations.

## Core idea

- **Legend:** the framework’s Go model, annotated with markers from `sail/legend`.
- **Chart:** the application’s Go struct, annotated with tags the Legend defines.
- **Navigator:** the compiled Legend grammar; `Build` interprets a Chart type,
  not the field values of a live instance.

## Install

Requires Go 1.27 or newer. After the first stable tag is published:

```sh
go get github.com/qlustra/sail@v1.0.0
```

## Quick start

In the [runnable quick start](testdata/publicapi/main.go), the Legend defines a
`job` family and a `name` Feature. A Chart field tagged `pipeline:"job" name:"compile"`
then contributes a job named `compile` to the populated model:

```go
navigator, err := sail.Compile[PipelineLegend]()
if err != nil {
    return err
}
model, err := navigator.Build[ReleaseChart]()
if err != nil {
    return err
}
fmt.Println(model.Jobs.Items[0].Name.Get())
```

From a Sail checkout, run the complete example in its separate module:

```sh
GOWORK=off go -C testdata/publicapi run .
# compile
```

## Project shape

- Root package `sail`: compilation, population, and navigation.
- `legend/`: grammar authoring markers.
- `testdata/publicapi/`: runnable quick start and consumer smoke test.

## Status

The v1 scope is Legend compilation and population, navigation with Survey,
Next and prepared Courses, and checked access to a recorded Chart field.

## Links

- [Sail documentation](https://github.com/Qlustra/sail-docs) — the usage manual.
- [Sail API](https://pkg.go.dev/github.com/qlustra/sail)
- [Legend markers](https://pkg.go.dev/github.com/qlustra/sail/legend)
