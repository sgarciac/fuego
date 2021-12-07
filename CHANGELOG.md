# Changelog

## [0.30.0]
### Fixes
 -Fixes bug in parsing. Breaking change, operations in, array-contains and
  array-contains-any are now [in], [array-contains] and [array-contains-any].

## [0.21.1]
### Added
- Documentation about snap.

## [0.21.0]
### Added
- Support for 'reference' data type.

## [0.20.1]
### Changed
- Remove --batch option from documentation.

## [0.20.0]
### Changed
- Input JSON is now interpreted using the semantics described in the TYPES.md
  file, which supports a larger set of firestore types.
- Documents output *can* be demanded to be transformed to the same notation, in
  order to have extra typing information.
- All documents are not displayed using the full document snapshot data
  information.
- Document snapshot data now includes the full path of the ducument.
- Several documentation improvements.

## [0.15.0]
### Changed
- Changes arguments for getall and deleteall commands.
- Adds documentation for creating releases.

## [0.14.0]
### Changed
- Add delete all operation


## [0.13.0]
### Changed
- Add get all operation

## [0.12.0]
### Changed
- Bug fixes

## [0.11.0]
### Added
- Support for document paths
- Support for group queries

## [0.10.0]
### Added
- Update firestore module.

## [0.9.2]
### Added
- fixed set merge

## [0.9.0]
### Added
- support update command (set-merge)


## [0.8.0]
### Added
- support array operations in, array-contains, array-contains-any

## [0.7.0]
### Added
- delete command

## [0.6.1]
### Added
- Documentation and dependencies

## [0.6.0]
### Added
- support firestore emulator

## [0.5.0]
### Added
- batched queries
- support complex property names in all commands
- support for multiple orderby's parameters 

## [0.4.0]
### Added
- collections command
- select flag in queries

## [0.3.0]
### Added
- Support arbitrary strings as field paths in queries

## [0.2.1]
### Added
- Support for timestamp values in queries
- Support for timestamp in writing operations operation

## [0.1.0] - 2019-02-27
### Added
- add,set,get and query commands
- README


