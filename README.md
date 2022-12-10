# Tower-Go

[![Build Status](https://drone.tigor.web.id/api/badges/tigorlazuardi/tower/status.svg)](https://drone.tigor.web.id/tigorlazuardi/tower)
[![Tower Test Status](https://minio.tigor.web.id/build-badges/tower/dist/tower-tests.svg)](https://drone.tigor.web.id/tigorlazuardi/tower)
[![TowerHTTP Test Status](https://minio.tigor.web.id/build-badges/tower/dist/towerhttp-tests.svg)](https://drone.tigor.web.id/tigorlazuardi/tower)

Note: this readme is still on WIP.

## Overview

Tower is an _**opinionated**_ Error, Logging, and Notification _framework_ for Go.

### TL; DR

This _framework_ will bring benefit to you when you want following features:

1. You want an error message that will help you track the origin of the error, regardless how lazy it's implemented.\*
2. You want a unified interface on how to handle errors. So if you want to change the output targets, the changes to your business code is small or even nothing at all.
3. You want control how an error is built.\*
4. You want easy assertion to search for Error.

\*This assumes you use the built-in implementations. Other implementations may have different results.

---

Tower takes huge inspiration from how HTTP Request is designed. HTTP Requests, no matter how different the implementation
is, it always have the same interface. Machines, Languages, etc. can interact with each other because the interface is
the same. Tower aims to be like that, One interface for the users, but the implementations can be wholly different.

It means that the user of the library can manage how the error will look, serialized, etc, without touching the business
part of the code.

## Tower is Designed for Working in a Team

Working in a team means you will not fully knows to the last detail on how your service works. Sure, you probably know
by heart about your part, but how about the other members in the team knowledge on your work?

**Code Review?** Sure, it will bring the idea of how the software works to you and bring you up to speed, but will you
know how it works? How about a month later? Still remember? Unlikely. Heck, you probably already forgot what code you
already work on by the next week. You will remember the rough idea about how the code works, but to the last detail like
what Mutex you use, which errors are handled, on which line, etc? Very unlikely unless you have eiditic memory.

Tower encourages users to bring `context` (like data, input, or output) with the error. While this is definitely not a new
concept, Tower does it in a way that allows that data processed further. Like for example, if I want my error to be
displayed in JSON, won't it be the best if the `context` also appears in JSON form? Querying such structured error would
be much better than mere stringified values.

## Tower Tries to Limit the Information to Relevant Ones

## Examples

TODO! (sorry! since the code is still in it's volatile stage, the signatures may change very rapidly. So it's best not to give examples for now).

# Afterwords

The library API design draws heavy inspiration from [mongodb-go](https://github.com/mongodb/mongo-go-driver) driver designs. Using Options that are split again into "Group" like options, just because of the sheer number of options available. `Tower`, while a lot smaller in scale, also needs such kind of flexibility. Hence the similar `Options` design.
