# Changelog

## [0.3.1](https://github.com/tclavelloux/promy-event-bus/compare/v0.3.0...v0.3.1) (2025-11-27)


### Features

* **deploy:** add Railway deployment for Redis service ([93d02a6](https://github.com/tclavelloux/promy-event-bus/commit/93d02a6062b52e8368b9debd2d97cdd1e0fb5a9f))


### Documentation

* **deploy:** add Railway deployment guide ([1fc959c](https://github.com/tclavelloux/promy-event-bus/commit/1fc959c7cbb57ec463c15a2cf963b596247eec2e))

## [0.3.0](https://github.com/tclavelloux/promy-event-bus/compare/v0.2.1...v0.3.0) (2025-11-25)


### ⚠ BREAKING CHANGES

* **events:** NewPromotionCreatedEvent signature changed from (promotionID, promotionName, distributorID, productTypeID string, dates []string, price float64, imageURL string) to (promotionID, promotionName, distributorID string, price float64, productTypeID *string, dates *[]string, imageURL *string, originalPrice *float64)

### Features

* **events:** add nullable fields to PromotionCreatedEvent ([f2c4ec7](https://github.com/tclavelloux/promy-event-bus/commit/f2c4ec7c74708298770b504d44334dbd8f72a77f))
* **pkg:** add ptr package for nullable field helpers ([adcc02b](https://github.com/tclavelloux/promy-event-bus/commit/adcc02b34b7f7db0333a6454cb2b8acebacc080e))


### Bug Fixes

* **ci:** exclude component name from Git tags ([b047167](https://github.com/tclavelloux/promy-event-bus/commit/b047167d4e24d2688cf51fb688f7b1ffe7845bf3))


### Tests

* **redis:** update tests for nullable event fields ([24b0853](https://github.com/tclavelloux/promy-event-bus/commit/24b0853eb756bee5c7b43378f12f2bf469146e8d))


### Documentation

* **examples:** update publisher example for nullable event fields ([ca8ebcd](https://github.com/tclavelloux/promy-event-bus/commit/ca8ebcdf6f3e6c53a06cae5ca33c3f67a987128b))

## [0.2.1](https://github.com/tclavelloux/promy-event-bus/compare/promy-event-bus-v0.2.0...promy-event-bus-v0.2.1) (2025-11-20)


### Features

* **events:** add validation tags to user events ([b54e4af](https://github.com/tclavelloux/promy-event-bus/commit/b54e4af1714676de91e4355535b31e1e9b3818ac))
* **events:** event schemas ([1dcb760](https://github.com/tclavelloux/promy-event-bus/commit/1dcb7609cf666d40ef18ed6e819de46233565699))
* **events:** event schemas for product, promotion and user preferences updates ([bdf5064](https://github.com/tclavelloux/promy-event-bus/commit/bdf50644f5a2364dbc9fe1c6c3f123dba28233e9))
* **redis:** integrate struct validation in publisher ([5475650](https://github.com/tclavelloux/promy-event-bus/commit/547565049c98479e24a1cb26f5e45f0cc45028a5))


### Bug Fixes

* **Github Actions:** fix release please manifest ([6a93f36](https://github.com/tclavelloux/promy-event-bus/commit/6a93f36a117a4d84e099f488cd18310240c8e5cd))


### Code Refactoring

* **eventbus:** remove legacy JSON schema validator ([3c35fd0](https://github.com/tclavelloux/promy-event-bus/commit/3c35fd00f1c3b6872748d82ed53dfe896e1af328))
* **promotion:** change categoryID for productTypeID ([04dc4de](https://github.com/tclavelloux/promy-event-bus/commit/04dc4decb4bfec3f7236aff6708290b16d8a96db))


### Build System

* **deps:** add go-playground/validator/v10 for struct validation ([bda0a48](https://github.com/tclavelloux/promy-event-bus/commit/bda0a487ee327425a32013aa291c5dbebc079907))
* **workflow:** add git workflow and release management ([45a5825](https://github.com/tclavelloux/promy-event-bus/commit/45a5825e3347a8b508618ad53fc495ccaac3560f))
* **workflow:** add git workflow and release management ([6dc1f8d](https://github.com/tclavelloux/promy-event-bus/commit/6dc1f8d0a713015b6ff7fac25cdd01748e55b774))


### Miscellaneous

* **main:** release 0.2.0 ([7b3ba83](https://github.com/tclavelloux/promy-event-bus/commit/7b3ba833f12ee397d278a30924d0247b2f3416e7))
* **main:** release 0.2.0 ([0f1a5e8](https://github.com/tclavelloux/promy-event-bus/commit/0f1a5e8268ab12ca13217ad43e136c78dbaccc3a))

## [0.2.0](https://github.com/tclavelloux/promy-event-bus/compare/v0.1.0...v0.2.0) (2025-11-19)


### Features

* **events:** add validation tags to user events ([b54e4af](https://github.com/tclavelloux/promy-event-bus/commit/b54e4af1714676de91e4355535b31e1e9b3818ac))
* **events:** event schemas ([1dcb760](https://github.com/tclavelloux/promy-event-bus/commit/1dcb7609cf666d40ef18ed6e819de46233565699))
* **events:** event schemas for product, promotion and user preferences updates ([bdf5064](https://github.com/tclavelloux/promy-event-bus/commit/bdf50644f5a2364dbc9fe1c6c3f123dba28233e9))
* **redis:** integrate struct validation in publisher ([5475650](https://github.com/tclavelloux/promy-event-bus/commit/547565049c98479e24a1cb26f5e45f0cc45028a5))
