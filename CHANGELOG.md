<a name="unreleased"></a>
## [Unreleased]

### Bug Fixes
- **caller:** fix wrong caller location for Drone ci
- **drone:** fix wrong makefile command
- **drone:** fix escapes
- **drone:** fix badge config
- **drone:** fix wrong makefile command
- **drone:** fix wrong discord image
- **drone:** fix wrong cache config
- **drone:** fix wrong cache config
- **drone:** fix wrong cache config
- **drone:** fix escapes
- **error-node-test:** changed test requirement to point to current file only
- **tower-discord:** better summary output
- **towererror-WriteError:** fix duplicate output

### Code Refactoring
- **caller:** caller is now an interface

### Docs
- update readme
- **drone:** added steps to build badges dynamically
- **tower-hints:** comments to satisfy lints

### Features
- **caller:** added missing methods
- **client-logger:** update client logger
- **discord:** better embed structures
- **discord:** added multipart uploads
- **drone:** update config
- **drone:** update config
- **drone:** added drone ci
- **drone:** config now ensures build is always cached
- **drone:** removed pull directive from save-cache and flush
- **drone:** added discord notifications
- **drone-discord:** removed message
- **drone-discord:** added message
- **entry:** added json marshaler support
- **error:** added deduped json marshaler support
- **error_node:** indent now set to 3 spaces
- **http-client:** wip update
- **implError:** exported implError as ErrorNode
- **query:** added bottom error query
- **query:** added CollectErrors query
- **respond-error:** tested common pattern
- **tower:** added code block marshaler pattern
- **tower-discord:** implemented file upload native discord
- **tower-discord:** added thread id in metadata
- **tower-http:** unit tested respond ok
- **tower-http:** major bug fixes with compressions
- **tower-http:** added unit test
- **tower-http:** added logging middleware to respond error
- **tower-http:** added logging middleware
- **tower-http:** added respond body logger
- **tower-http:** added client logger
- **tower-http:** added response capturer
- **tower-http:** unit tested respond ok with http no body
- **tower-http:** added unit test
- **tower-http:** update error body
- **tower-http:** more unit test to Respond Ok
- **tower-http:** unit tested global respond
- **tower-http:** update
- **tower-http:** unit tested major refactor
- **tower-http-exported:** added env to skip test global instances
- **tower-http-gzip:** skip compression on data too small
- **tower-http-gzip:** skip compression on data too small
- **travis-ci:** added travis ci to run tests

### Miscellaneous
- synchronize go mod with new tag on tower

### Test
- **respond-error:** added tower error pattern test
- **tower-discord:** tested hooks
- **tower-http:** more test cases
- **tower-http:** start testing respond stream
- **tower-http:** added more test cases
- **tower-http:** added more test cases
- **tower-http:** added more test for RespondError
- **tower-http-respond-ok:** added mock compress error test

### Wip
- **tower-http:** refactor signature


<a name="v0.1.9"></a>
## [v0.1.9] - 2022-11-26
### Miscellaneous
- sync again


<a name="v0.1.8"></a>
## [v0.1.8] - 2022-11-26

<a name="v0.1.7"></a>
## [v0.1.7] - 2022-11-26
### Miscellaneous
- FIX THOSE TAGS WTF GO PROXY!?


<a name="v0.1.5"></a>
## [v0.1.5] - 2022-11-26
### Miscellaneous
- cleaned up comments


<a name="v0.1.4"></a>
## [v0.1.4] - 2022-11-26

<a name="v0.1.3"></a>
## [v0.1.3] - 2022-11-26
### Bug Fixes
- **queue:** scrapped nasty bug for queue logic and used native golang implementation instead.

### Features
- **discord:** major update and fixes
- **discord:** separated upload file flow between bucket and discord
- **discord:** add further implementations
- **discord:** added hook
- **tower:** added json marshaler to error and caller
- **tower-http:** added client logger

### Miscellaneous
- explicit ignore on lints

### Tag
- added new tag

### Wip
- **discord:** post message


<a name="v0.1.2"></a>
## [v0.1.2] - 2022-11-25
### Bug Fixes
- **responder:** fix bad refactorings
- **responder:** fix bad refactoring

### Code Refactoring
- **responder:** rename option
- **responder:** simplified the api
- **responder:** moved RespondStream to its own file

### Features
- **discord:** further updates
- **discord:** added setters for discord bot
- **discord:** added embed building
- **responder:** added exported version
- **responder:** added error responder and constructor
- **responder:** http.NoBody now no longer set default status code to http.StatusNoContent
- **responder:** added special handling for nil and http.NoBody

### Miscellaneous
- cleanup go mods


<a name="v0.1.1"></a>
## [v0.1.1] - 2022-11-23
### Bug Fixes
- **error:** fix missing code hints
- **fields:** fix wrong logic on writing empty json
- **queue:** go dependency changed to 1.18
- **slackbot:** missing upload call command

### Code Refactoring
- renamed GzipCompressor to GzipCompression for consistency
- removed error from body transform
- renamed compressor to compression
- compression Compress method now returns 3 values
- moved caller from hints file to caller file
- removed body code hint because it's purpose is ambiguous
- removed constraints lib
- **GetCaller:** changed signature to return zero value instead of with OK on failure to capture frame.
- **LineWriter:** renamed methods for better readability for implementers
- **discord:** added block scope to set explicit where the variable pointers pointed to
- **lefthook:** lefthook now uses make command
- **line-writer:** reduced requirement for LineWriter to merely io.Writer
- **stream-compression:** split the definition for streams for easier api usage
- **towerhttp:** moved Respond method to its own file
- **towerslack:** updated internal data signature

### Docs
- update code documentations to follow better standard
- update docs
- **changelog:** changed the format
- **tower:** update message context docs
- **tower-query:** update docs
- **wrap:** update the docs on wrap

### Features
- renamed Option to MessageParameter
- added exported functions
- added auto changelog
- update options
- updated implementations for todos
- update messenger spec
- general update
- major update
- added messenger
- general update
- more updates
- **Messenger:** changed signature so ctx can be modified
- **blocks:** added more blocks
- **blocks:** added section block
- **bucket:** added bucket interface
- **cache:** added cacher interface
- **commitlint:** added commitlint integration
- **context-builder:** added context builder impls
- **discord:** updated discord element
- **discord:** added send
- **discord:** update discord
- **discord:** update discord
- **discord:** file uploads are now markdown
- **discord:** thread-id is generated
- **discord:** added summary buildre
- **discord:** added data embed builder
- **entry:** added Entry implementations
- **error:** properly implemented defaultErrorGenerator
- **fields:** major fields update
- **gitignore:** goland .idead folder is now gitignored
- **message-context:** added implmentations
- **message-option:** simplified the api
- **option:** added status code override option
- **query:** added query methods
- **queue:** added queue
- **queue:** uses lock free queue algorithm instead of two lock queue since it's faster
- **respond-stream:** support for tower.HTTPCodeHint interface check
- **respond-stream:** support for tower.HTTPCodeHint interface check
- **slack:** added slack
- **slack:** update
- **slack:** added post message
- **slack:** implemented towerslack handle message
- **slack:** added dynamic build key
- **slackbot:** implemented call to file attachments
- **tower:** added wrap method
- **tower-http:** added towerhttp library
- **towerhttp:** added compressor and respond method
- **towerhttp:** added RequestContext logging
- **towerslack:** update documentations.
- **towerslack:** added constructor and template builder.
- **towerslack:** start building template
- **towerzap:** added towerzap implementations
- **workspace:** now uses workspace to separate dependencies
- **writer:** added writer implementation

### Fix
- **tower-query:** Fix typo on GetCodeHint returned value

### Miscellaneous
- renamed Logger to NoopLogger in comment
- format
- removed commitlint from githooks
- rename interface to a better name
- cleanup uneeded sum files
- update gitignore
- comment linter for wip
- format comments


<a name="v0.1.0"></a>
## v0.1.0 - 2022-10-20
### Docs
- added docs support


[Unreleased]: https://github.com/tigorlazuardi/tower/compare/v0.1.9...HEAD
[v0.1.9]: https://github.com/tigorlazuardi/tower/compare/v0.1.8...v0.1.9
[v0.1.8]: https://github.com/tigorlazuardi/tower/compare/v0.1.7...v0.1.8
[v0.1.7]: https://github.com/tigorlazuardi/tower/compare/v0.1.5...v0.1.7
[v0.1.5]: https://github.com/tigorlazuardi/tower/compare/v0.1.4...v0.1.5
[v0.1.4]: https://github.com/tigorlazuardi/tower/compare/v0.1.3...v0.1.4
[v0.1.3]: https://github.com/tigorlazuardi/tower/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/tigorlazuardi/tower/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/tigorlazuardi/tower/compare/v0.1.0...v0.1.1
