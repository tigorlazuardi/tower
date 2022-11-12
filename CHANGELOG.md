<a name="unreleased"></a>
## [Unreleased]

### Bug Fixes
- **error:** fix missing code hints
- **queue:** go dependency changed to 1.18

### Code Refactoring
- removed body code hint because it's purpose is ambiguous
- removed constraints lib
- moved caller from hints file to caller file
- **lefthook:** lefthook now uses make command

### Docs
- update code documentations to follow better standard
- update docs
- **changelog:** changed the format
- **wrap:** update the docs on wrap

### Features
- added exported functions
- update options
- added messenger
- update messenger spec
- more updates
- updated implementations for todos
- general update
- general update
- added auto changelog
- renamed Option to MessageParameter
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
- **message-context:** added implmentations
- **message-option:** simplified the api
- **query:** added query methods
- **queue:** added queue
- **queue:** uses lock free queue algorithm instead of two lock queue since it's faster
- **slack:** implemented towerslack handle message
- **slack:** added dynamic build key
- **slack:** added slack
- **slack:** added post message
- **slack:** update
- **tower:** added wrap method
- **towerslack:** start building template
- **towerslack:** added constructor and template builder.
- **towerzap:** added towerzap implementations
- **workspace:** now uses workspace to separate dependencies
- **writer:** added writer implementation

### Miscellaneous
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
