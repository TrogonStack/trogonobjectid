# Changelog

## [0.4.0](https://github.com/TrogonStack/trogonobjectid/compare/v0.3.0...v0.4.0) (2026-10-22)


### ⚠ BREAKING CHANGES

* narrow unions with UnionAs and name member lists Types ([#7](https://github.com/TrogonStack/trogonobjectid/issues/7))
* name object types and separators, and check definitions early ([#4](https://github.com/TrogonStack/trogonobjectid/issues/4))
* take a base Definition in objectidproto and cache definitions per type ([#3](https://github.com/TrogonStack/trogonobjectid/issues/3))

### Features

* Accept earlier prefixes with Definition.Aliases ([#8](https://github.com/TrogonStack/trogonobjectid/issues/8)) ([542fc90](https://github.com/TrogonStack/trogonobjectid/commit/542fc90cb33936cccfbc86926939430a74665600))
* Add objectidtest for deterministic IDs in tests ([#6](https://github.com/TrogonStack/trogonobjectid/issues/6)) ([869bcb9](https://github.com/TrogonStack/trogonobjectid/commit/869bcb927e3ea9fd64b61198dc685355e885bb3e))
* Build IDs from UUIDs and scan native UUID values ([#5](https://github.com/TrogonStack/trogonobjectid/issues/5)) ([f4bf2de](https://github.com/TrogonStack/trogonobjectid/commit/f4bf2de9952684160caf593b80bc51ec37b390e5))
* Name object types and separators, and check definitions early ([#4](https://github.com/TrogonStack/trogonobjectid/issues/4)) ([3856a56](https://github.com/TrogonStack/trogonobjectid/commit/3856a56890a720102f8e6593c07df5adb3da662e))
* Narrow unions with UnionAs and name member lists Types ([#7](https://github.com/TrogonStack/trogonobjectid/issues/7)) ([91b78b0](https://github.com/TrogonStack/trogonobjectid/commit/91b78b06e19c87932d079539e183ff218ff3d160))
* Take a base Definition in objectidproto and cache definitions per type ([#3](https://github.com/TrogonStack/trogonobjectid/issues/3)) ([8d466df](https://github.com/TrogonStack/trogonobjectid/commit/8d466df0efe5a13c146ef20c38c788add0925b12))

## [0.3.0](https://github.com/TrogonStack/trogonobjectid/compare/v0.2.0...v0.3.0) (2026-10-22)


### ⚠ BREAKING CHANGES

* accept only canonical UUIDs and drop Normalize ([#12](https://github.com/TrogonStack/trogonobjectid/issues/12))

### Features

* Accept only canonical UUIDs and drop Normalize ([#12](https://github.com/TrogonStack/trogonobjectid/issues/12)) ([1e35add](https://github.com/TrogonStack/trogonobjectid/commit/1e35add8957592fc51ab253a290879ff3da071f3))

## [0.2.0](https://github.com/TrogonStack/trogonobjectid/compare/v0.1.0...v0.2.0) (2026-10-22)


### Features

* Normalize bare values and encode IDs for Spanner ([#2](https://github.com/TrogonStack/trogonobjectid/issues/2)) ([40f3ff3](https://github.com/TrogonStack/trogonobjectid/commit/40f3ff3fadd919059d0ccc9e48f69252c4a8f795))

## 0.1.0 (2026-10-15)


### Features

* Add type-safe prefixed object IDs ([#1](https://github.com/TrogonStack/trogonobjectid/issues/1)) ([ac47c73](https://github.com/TrogonStack/trogonobjectid/commit/ac47c738297b581543d74db9322dd74a824d6b8e))
* Add unions of object ID types ([#2](https://github.com/TrogonStack/trogonobjectid/issues/2)) ([8f77364](https://github.com/TrogonStack/trogonobjectid/commit/8f77364b9fc4dcadfd1483dc98f69704ffba95f5))
* Read object ID definitions from proto annotations ([#3](https://github.com/TrogonStack/trogonobjectid/issues/3)) ([a439749](https://github.com/TrogonStack/trogonobjectid/commit/a4397498e9024c71c940c8a4d247fa799822c253))
