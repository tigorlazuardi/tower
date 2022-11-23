<a name="unreleased"></a>
## [Unreleased]

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
- update options
- update messenger spec
- added exported functions
- more updates
- added auto changelog
- renamed Option to MessageParameter
- major update
- general update
- general update
- added messenger
- updated implementations for todos
- **Messenger:** changed signature so ctx can be modified
- **blocks:** added section block
- **blocks:** added more blocks
- **bucket:** added bucket interface
- **cache:** added cacher interface
- **commitlint:** added commitlint integration
- **context-builder:** added context builder impls
- **discord:** updated discord element
- **discord:** added data embed builder
- **discord:** update discord
- **discord:** update discord
- **discord:** added send
- **discord:** thread-id is generated
- **discord:** added summary buildre
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
- **slack:** added dynamic build key
- **slack:** added post message
- **slack:** update
- **slack:** added slack
- **slack:** implemented towerslack handle message
- **slackbot:** implemented call to file attachments
- **tower:** added wrap method
- **tower-http:** added towerhttp library
- **towerhttp:** added RequestContext logging
- **towerhttp:** added compressor and respond method
- **towerslack:** start building template
- **towerslack:** update documentations.
- **towerslack:** added constructor and template builder.
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


[Unreleased]: https://github.com/tigorlazuardi/tower/compare/v0.1.0...HEAD
