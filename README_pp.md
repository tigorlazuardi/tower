# Tower-Go

[![Build Status](https://drone.tigor.web.id/api/badges/tigorlazuardi/tower/status.svg)](https://drone.tigor.web.id/tigorlazuardi/tower)
[![Tower Test Status](https://minio.tigor.web.id/build-badges/tower/dist/tower-tests.svg)](https://drone.tigor.web.id/tigorlazuardi/tower)
[![TowerHTTP Test Status](https://minio.tigor.web.id/build-badges/tower/dist/towerhttp-tests.svg)](https://drone.tigor.web.id/tigorlazuardi/tower)

Note: this readme is still on WIP.

## Overview

Tower is an _**opinionated**_ Error, Logging, and Notification _framework_ for Go.

Tower's main goal is to improve developer experience when handling errors and logging.

Tower's main goal is to provide a common API interface for error handling, logging, and notification. It also aims to
provide more information about the error, such as where the error occurred. It also optionally provides a way to enrich
the error with additional information, such as a message, data, error code, and so on.

Tower does not stop there, it goes one step further by providing a way to log and send the error to a notification in
one single flow.

It basically turns the flow from this:

```go
func foo() error {
    _, err := strconv.Atoi("foo")
    if err != nil {
        err := fmt.Errorf("failed to convert string to int: %w", err)
        log.Println(err)
        notify(ctx, err)
        return err
    }
}
```

Into this:

```go
func foo() error {
    _, err := strconv.Atoi("foo")
    if err != nil {
        return tower.Wrap(err).Message("failed to convert string to int").Log(ctx).Notify(ctx) // Notify and Log in one single flow
        // return tower.Wrap(err).Message("failed to convert string to int").Log(ctx)  <-- if you just want to log
        // return tower.Wrap(err).Message("failed to convert string to int").Notify(ctx)  <-- if you want to send notification.
        // return tower.Wrap(err).Message("failed to convert string to int").Freeze() <-- if you just want to enrich the error.
        // return tower.WrapFreeze(err, "failed to convert string to int") <-- short hand for above.
    }
}
```

I could already hear you saying "Hey I still have to write that much, what gives?"

There are already a lot of things happening behind the scenes, such as:

- When you call `tower.Wrap(err)`, it will automatically enrich the error with the location of the Wrap() caller, and
  the error code. The returned type is an `ErrorBuilder`, which means it can be chained with other methods to enrich the
  error
  further.
- When you call `.Log(ctx)`, it will make Tower to look at its own `tower.Logger` implementor, and sends the enriched
  Error to the logger.
- When you call `.Notify(ctx)`, it will make Tower to look at its own `tower.Messenger` implementors, and sends the
  enriched Error to the messengers.
- When you call `.Freeze()`, it will make Tower transforms the mutable `ErrorBuilder` into an immutable `Error`.

There are already obvious benefits from this snippet alone. Logging and Sending Notifications are decoupled from the
business logic, and the business logic is now more readable. You can also easily change or add more logging and
notification to Tower itself and the changes will be reflected without modifying the business logic.

If you are working with a team, for them, it's already obvious that you already log and send notification from here when
error happens. More often than not for them, it's enough. They most likely don't want to know how you log and send
notification, they just want to know that you do. Hell, future you probably just want to know that your current self do
it and don't care about the details.

While the first snippet fulfills above conditions, the relationship between logging and business is strictly coupled and
makes refactoring a chore.

While these things can be considered "little things", they do add up and make a huge difference in the long run.

## Logger, Messenger, Entry, Error

TODO: write this section

# Afterwords

The library API design draws heavy inspiration from [mongodb-go](https://github.com/mongodb/mongo-go-driver) driver
designs. Using Options that are split again into "Group" like options, just because of the sheer number of options
available. `Tower`, while a lot smaller in scale, also needs such kind of flexibility. Hence the similar `Options`
design.
