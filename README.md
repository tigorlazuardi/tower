# Tower-Go [![build-status-svg]][drone-dashboard]

## Coverages

| Module Type  | Coverage                                                                                                     |
| ------------ | ------------------------------------------------------------------------------------------------------------ |
| Root         | ![tower-svg]                                                                                                 |
| Integrations | ![cache-svg] <br /> ![bucket-svg]                                                                            |
| Extensions   | ![towerhttp-svg] <br /> ![towerzap-svg] <br /> ![towerdiscord-svg]                                           |
| Submodules   | ![memcache-svg] <br /> ![goredis-v8-svg] <br /> ![goredis-v9-svg] <br /> ![minio-v7-svg] <br /> ![s3-v2-svg] |
| Utilities    | ![pool-svg] <br /> ![queue-svg] <br /> ![loader-svg]                                                         |

[build-status-svg]: https://drone.tigor.web.id/api/badges/tigorlazuardi/tower/status.svg 'Build Status'
[tower-svg]: https://minio.tigor.web.id/badges/tower/tower.svg? 'Tower Test Coverage'
[towerhttp-svg]: https://minio.tigor.web.id/badges/tower/towerhttp.svg? 'TowerHTTP Test Coverage'
[towerzap-svg]: https://minio.tigor.web.id/badges/tower/towerzap.svg? 'Towerzap Test Coverage'
[towerdiscord-svg]: https://minio.tigor.web.id/badges/tower/towerdiscord.svg? 'Towerdiscord Test Coverage'
[cache-svg]: https://minio.tigor.web.id/badges/tower/cache.svg? 'Cache Test Coverage'
[bucket-svg]: https://minio.tigor.web.id/badges/tower/bucket.svg? 'Cache Test Coverage'
[memcache-svg]: https://minio.tigor.web.id/badges/tower/cache/gomemcache.svg? 'Go-Memcached Test Coverage'
[goredis-v8-svg]: https://minio.tigor.web.id/badges/tower/cache/goredis/v8.svg? 'GoRedis V8 Test Coverage'
[goredis-v9-svg]: https://minio.tigor.web.id/badges/tower/cache/goredis/v9.svg? 'GoRedis V9 Test Coverage'
[minio-v7-svg]: https://minio.tigor.web.id/badges/tower/bucket/minio/v7.svg 'Minio V7 Test Coverage'
[s3-v2-svg]: https://minio.tigor.web.id/badges/tower/bucket/s3/v2.svg 'S3 V2 Test Coverage'
[pool-svg]: https://minio.tigor.web.id/badges/tower/pool.svg? 'Pool Test Coverage'
[queue-svg]: https://minio.tigor.web.id/badges/tower/queue.svg? 'Queue Test Coverage'
[loader-svg]: https://minio.tigor.web.id/badges/tower/loader.svg? 'Loader Test Coverage'
[drone-dashboard]: https://drone.tigor.web.id/tigorlazuardi/tower 'Drone Dashboard'

Note: this readme is still on WIP.

# Overview

Tower is an _**opinionated**_ Error, Logging, and Notification _framework_ for Go.

Tower's main goal is to improve developer experience when handling errors and logging.

Tower does so by providing a common API interface for error handling, logging, and notification. It also aims to provide
more information about the error, such as where the error occurred or the context of how it happens. It also optionally
provides a way to enrich the error with additional information, such as a message, data, error code, and so on.

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

-   When you call `tower.Wrap(err)`, it will automatically enrich the error with the location of the Wrap() caller,
    settings the log level into `tower.ErrorLevel` and fill the error code.
-   When you call `.Log(ctx)`, it will make Tower to look at its own `tower.Logger` implementor, and sends the enriched
    Error to the logger.
-   When you call `.Notify(ctx)`, it will make Tower to look at its own `tower.Messenger` implementors, and sends the
    enriched Error to the messengers.
-   When you call `.Freeze()`, it will make Tower transforms the mutable `ErrorBuilder` into an immutable `Error`.

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

# Technical Details

## Error, ErrorBuilder, Entry, EntryBuilder, Logger, Messenger

All these 6 entities are the main Types you have to know when using Tower. As a user, You don't have to know how they
exactly works, but knowing the idea behind them will let you get the most out of `Tower` framework.

### ErrorBuilder

[tower.ErrorBuilder](https://pkg.go.dev/github.com/tigorlazuardi/tower#ErrorBuilder) is the entry point to handle errors
with Tower. Whenever [tower.Wrap](https://pkg.go.dev/github.com/tigorlazuardi/tower#Wrap) function is called, this type
is returned.

Note that `tower.ErrorBuilder` itself **_does not implement `error`_**. It holds temporary and mutable values until one
of these three methods were called:

1. `.Freeze()`, turns the `tower.ErrorBuilder` into `tower.Error`, which implements `error`.
2. `.Log(context.Context)`, implicitly calls `.Freeze()` and then calls `.Log(context.Context)` of the `tower.Error`.
3. `.Notify(context.Context, ...MessageOption)`, implicitly calls `.Freeze()`, and then calls `.Notify`of the
   `tower.Error`

Realistic example of using the `tower.ErrorBuilder`.

```go
func foo(s string) error {
    _, err := strconv.Atoi(s)
    if err != nil {
        return tower.
            Wrap(err).
            Message("strconv: failed to convert string to int").
            Code(400).
            Context(tower.Fields{
                "input": s,
            }).
            Log(ctx). // calls .Freeze() implicitly, turning into tower.Error, calls the Log method of tower.Error, then
                      // return the tower.Error
            Notify(ctx)
            // There are more API in tower.ErrorBuilder, consult the docs for those.
    }
}
```

### Error

[tower.Error](https://pkg.go.dev/github.com/tigorlazuardi/tower#Error) is an extension to the native golang's `error`
type. `tower.Error` is a crystallized form of values from

# Afterwords

The library API design draws heavy inspiration from [mongodb-go](https://github.com/mongodb/mongo-go-driver) driver
designs. Using Options that are split again into "Group" like options, just because of the sheer number of options
available. `Tower`, while a lot smaller in scale, also needs such kind of flexibility. Hence the similar `Options`
design.
