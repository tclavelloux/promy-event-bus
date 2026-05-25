# Changelog

## [0.4.0](https://github.com/tclavelloux/promy-event-bus/compare/v0.3.1...v0.4.0) (2026-05-25)


### ⚠ BREAKING CHANGES

* The events/ package (event structs, type constants) is removed. Each producing service now owns its own event definitions in internal/events/. Stream name constants move to the streams/ package.

### Features

* **ci:** add registry validation script and GitHub Actions workflow ([7fe60a1](https://github.com/tclavelloux/promy-event-bus/commit/7fe60a19dd4244044cb7e19328faf9ba738f802f))
* create event schema registry with seed events ([7bbd3ae](https://github.com/tclavelloux/promy-event-bus/commit/7bbd3aed1ee9c2cb5bacba1e908f613fd4c475cc))
* **dlq:** add inspect and replay CLI tool ([5e0c7bf](https://github.com/tclavelloux/promy-event-bus/commit/5e0c7bf0b687bd6851349fc8f5f6f5b03d3cd37c))
* **eventbus:** add Data() string to Event interface ([1d2982a](https://github.com/tclavelloux/promy-event-bus/commit/1d2982a2dffac5fa419ba1aaa09e341fc2e43287))
* **eventbus:** add DLQ entry helper and subscriber routing ([bebeea4](https://github.com/tclavelloux/promy-event-bus/commit/bebeea40b4ddf6998539b60cbdbb2ab3ae6440d0))
* **streams:** add StreamSubscriptions, StreamIdentifications, and StreamDLQ constants ([4548cb9](https://github.com/tclavelloux/promy-event-bus/commit/4548cb91e76552e755b0746cefedecd681ef745e))


### Code Refactoring

* **redis:** extract metadata and payload field names as constants ([6fa0c08](https://github.com/tclavelloux/promy-event-bus/commit/6fa0c08a4421b705743fe3fab224b616e311786c))
* remove events/ package, move stream constants to streams/ ([0489f74](https://github.com/tclavelloux/promy-event-bus/commit/0489f74a4c0b01fe681e79e47e5eb2eb9ecd7cc0))


### Build System

* **docker:** change Redis host port from 6379 to 6389 ([1cc5002](https://github.com/tclavelloux/promy-event-bus/commit/1cc50023a00603f48498b1a2b916b24d453d52c5))
* **make:** shorten docker compose target names ([23ed925](https://github.com/tclavelloux/promy-event-bus/commit/23ed9258a45b2c299910cfccc206c256f09a5797))


### Documentation

* add CLAUDE.md with codebase guidance for Claude Code ([e2bd077](https://github.com/tclavelloux/promy-event-bus/commit/e2bd077804220685ac7abd42df1d87a7f1af3791))
* add integration guide for downstream Yokai services ([33379d4](https://github.com/tclavelloux/promy-event-bus/commit/33379d42d3d3e8ebe1e3b90e8d378d023219588b))
* replace HOWTO with v2.1 ([5fde8a9](https://github.com/tclavelloux/promy-event-bus/commit/5fde8a9395311d5277c72b9864e7a5af2f641235))
* rewrite README to reflect current architecture ([b2a5189](https://github.com/tclavelloux/promy-event-bus/commit/b2a51890c754b99f36a3a59ed2fb205f1998f5e9))
* update README and HOWTO for DLQ routing ([6356271](https://github.com/tclavelloux/promy-event-bus/commit/63562717cafa6b928115df8786115b8a4986d41a))


### Miscellaneous

* **cursor:** remove tracked cursor rules and docs from repo ([6926110](https://github.com/tclavelloux/promy-event-bus/commit/69261107ae8b1a27af84c384df854d6876a30c1f))
* **cursor:** remove tracked cursor rules and docs from repo ([036215e](https://github.com/tclavelloux/promy-event-bus/commit/036215ef25250dc3f51e79233e552220926ef63d))
* **docs:** remove outdated cursor git workflow docs ([dc45743](https://github.com/tclavelloux/promy-event-bus/commit/dc457430cad2aa7855495e5d1e0d1cf823bb1fbb))
* **gh actions:** wire token ([363ef5b](https://github.com/tclavelloux/promy-event-bus/commit/363ef5b9550abdd908f9631d4710f8817dd25278))
* **gh actions:** wire token ([0100449](https://github.com/tclavelloux/promy-event-bus/commit/01004499732a860af61e6d5e09dfc95c68e146ba))
* port infrastructure and tooling adaptations ([f9a12cd](https://github.com/tclavelloux/promy-event-bus/commit/f9a12cdf9aedd01ae6af541112af2de32c1d1540))

## [0.3.1](https://github.com/tclavelloux/promy-event-bus/compare/v0.3.0...v0.3.1) (2025-12-13)


### Features

* **deploy:** add Railway deployment for Redis service ([93d02a6](https://github.com/tclavelloux/promy-event-bus/commit/93d02a6062b52e8368b9debd2d97cdd1e0fb5a9f))
* **events:** update promotion event date handling ([fb954e0](https://github.com/tclavelloux/promy-event-bus/commit/fb954e033ac6dc69cdd76376b984024dba0a0f00))
* **events:** update promotion event date handling ([2af8ee4](https://github.com/tclavelloux/promy-event-bus/commit/2af8ee4a95ea9908828d9f42e785076f03cb6a52))


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
