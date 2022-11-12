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
- update messenger spec
- renamed Option to MessageParameter
- more updates
- added auto changelog
- update options
- general update
- updated implementations for todos
- general update
- added messenger
- added exported functions
- **Messenger:** changed signature so ctx can be modified
- **blocks:** added section block
- **blocks:** added more blocks
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
- **slack:** update
- **slack:** added slack
- **slack:** added post message
- **slack:** implemented towerslack handle message
- **slack:** added dynamic build key
- **tower:** added wrap method
- **towerslack:** update documentations.
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
