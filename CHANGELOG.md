<a name="unreleased"></a>
## [Unreleased]

### Bug Fixes
- **queue:** go dependency changed to 1.18

### Code Refactoring
- removed constraints lib
- moved caller from hints file to caller file
- **lefthook:** lefthook now uses make command

### Docs
- update docs
- **changelog:** changed the format

### Features
- update options
- added exported functions
- added auto changelog
- updated implementations for todos
- renamed Option to MessageParameter
- added messenger
- update messenger spec
- general update
- general update
- more updates
- **Messenger:** changed signature so ctx can be modified
- **blocks:** added more blocks
- **blocks:** added section block
- **cache:** added cacher interface
- **commitlint:** added commitlint integration
- **discord:** updated discord element
- **entry:** added Entry implementations
- **error:** properly implemented defaultErrorGenerator
- **fields:** major fields update
- **message-option:** simplified the api
- **queue:** added queue
- **queue:** uses lock free queue algorithm instead of two lock queue since it's faster
- **slack:** implemented towerslack handle message
- **slack:** added dynamic build key
- **slack:** added slack
- **slack:** added post message
- **slack:** update
- **tower:** added wrap method
- **towerslack:** start building template
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
