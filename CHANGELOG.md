<a name="unreleased"></a>
## [Unreleased]

### Bug Fixes
- **error:** fix missing code hints
- **fields:** fix wrong logic on writing empty json
- **queue:** go dependency changed to 1.18
- **slackbot:** missing upload call command

### Code Refactoring
- removed error from body transform
- compression Compress method now returns 3 values
- renamed GzipCompressor to GzipCompression for consistency
- renamed compressor to compression
- moved caller from hints file to caller file
- removed body code hint because it's purpose is ambiguous
- removed constraints lib
- **GetCaller:** changed signature to return zero value instead of with OK on failure to capture frame.
- **LineWriter:** renamed methods for better readability for implementers
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
- added exported functions
- general update
- added messenger
- update messenger spec
- more updates
- major update
- general update
- update options
- renamed Option to MessageParameter
- added auto changelog
- updated implementations for todos
- **Messenger:** changed signature so ctx can be modified
- **blocks:** added more blocks
- **blocks:** added section block
- **bucket:** added bucket interface
- **cache:** added cacher interface
- **commitlint:** added commitlint integration
- **context-builder:** added context builder impls
- **discord:** updated discord element
- **entry:** added Entry implementations
- **error:** properly implemented defaultErrorGenerator
- **fields:** major fields update
- **gitignore:** goland .idead folder is now gitignored
- **message-context:** added implmentations
- **message-option:** simplified the api
- **option:** added status code override option
- **query:** added query methods
- **queue:** uses lock free queue algorithm instead of two lock queue since it's faster
- **queue:** added queue
- **slack:** added slack
- **slack:** added post message
- **slack:** update
- **slack:** added dynamic build key
- **slack:** implemented towerslack handle message
- **slackbot:** implemented call to file attachments
- **tower:** added wrap method
- **tower-http:** added towerhttp library
- **towerhttp:** added compressor and respond method
- **towerhttp:** added RequestContext logging
- **towerslack:** update documentations.
- **towerslack:** start building template
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
