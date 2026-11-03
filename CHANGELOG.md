# Changelog

## [0.5.1](https://github.com/tclavelloux/promy-event-bus/compare/v0.5.0...v0.5.1) (2026-10-31)


### Build System

* **deps:** follow governance hooks v1.9.0 (go-test-coverage 2.21.0) ([#81](https://github.com/tclavelloux/promy-event-bus/issues/81)) ([3b60d6a](https://github.com/tclavelloux/promy-event-bus/commit/3b60d6a80606540ba07b02f882b4cedb895c9836))


### Miscellaneous

* **claude:** enable promy plugin ([#80](https://github.com/tclavelloux/promy-event-bus/issues/80)) ([d4c5f6e](https://github.com/tclavelloux/promy-event-bus/commit/d4c5f6edce8efd0595329d40f42db15bf92bc30e))
* **dev:** add the shared promy-jaeger trace backend ([#77](https://github.com/tclavelloux/promy-event-bus/issues/77)) ([d0efa21](https://github.com/tclavelloux/promy-event-bus/commit/d0efa21050447504a7d9f6acd2a074ff71482883))

## [0.5.0](https://github.com/tclavelloux/promy-event-bus/compare/v0.4.2...v0.5.0) (2026-10-30)


### ⚠ BREAKING CHANGES

* **registry:** move product.identified to events:products and drop events:identifications ([#73](https://github.com/tclavelloux/promy-event-bus/issues/73))

### Features

* **observability:** propagate W3C trace context through Redis Streams and log dropped messages ([#76](https://github.com/tclavelloux/promy-event-bus/issues/76)) ([c9ac810](https://github.com/tclavelloux/promy-event-bus/commit/c9ac810e357aa823850fafc3ed0a679941bcc271))
* **subscriber:** report swallowed DLQ, requeue and ack errors through an optional callback ([#74](https://github.com/tclavelloux/promy-event-bus/issues/74)) ([ddcbe04](https://github.com/tclavelloux/promy-event-bus/commit/ddcbe043b78d584c68a153b6b2646582f44b909a))


### Bug Fixes

* **registry:** move product.identified to events:products and drop events:identifications ([#73](https://github.com/tclavelloux/promy-event-bus/issues/73)) ([4104238](https://github.com/tclavelloux/promy-event-bus/commit/410423842946898d4391f0ddb0b7cdbc6beb15c3))


### Build System

* **deps:** bump github.com/gabriel-vasile/mimetype from 1.4.15 to 1.4.16 in the gomod group across 1 directory ([#67](https://github.com/tclavelloux/promy-event-bus/issues/67)) ([3328eb7](https://github.com/tclavelloux/promy-event-bus/commit/3328eb76c6e83d073cde99630ed47063f2a18be5))


### Documentation

* align the DLQ contract and refresh README and HOWTO ([#72](https://github.com/tclavelloux/promy-event-bus/issues/72)) ([c231b9a](https://github.com/tclavelloux/promy-event-bus/commit/c231b9aa6b43f0e2eaeae4ab7d0547117d55769b))


### Miscellaneous

* **hooks:** bump governance hooks to v1.7.0 and go-test-coverage to v2.20.1 ([#70](https://github.com/tclavelloux/promy-event-bus/issues/70)) ([0f423bc](https://github.com/tclavelloux/promy-event-bus/commit/0f423bc8f3b18b1828d64b5da897e8b525b7f68c))
* **hooks:** run each pre-commit hook once per commit ([#68](https://github.com/tclavelloux/promy-event-bus/issues/68)) ([56ad61a](https://github.com/tclavelloux/promy-event-bus/commit/56ad61a9d79d34f721c45d8c6fac40193c39b5ea))

## [0.4.2](https://github.com/tclavelloux/promy-event-bus/compare/v0.4.1...v0.4.2) (2026-10-26)


### Bug Fixes

* **deps:** bump Go to 1.26.9 and pass the App secrets to vuln-fix ([#62](https://github.com/tclavelloux/promy-event-bus/issues/62)) ([13597f9](https://github.com/tclavelloux/promy-event-bus/commit/13597f98e44f281b7a90488ca914107b2cf705cf))
* **deps:** patch GO-2026-6809 ([#65](https://github.com/tclavelloux/promy-event-bus/issues/65)) ([fad4338](https://github.com/tclavelloux/promy-event-bus/commit/fad4338c587116ad8c6f0bd18a05079decfc17bb))
* **registry:** set the events:subscriptions owner to promy-subscription ([#51](https://github.com/tclavelloux/promy-event-bus/issues/51)) ([af9ac49](https://github.com/tclavelloux/promy-event-bus/commit/af9ac49b3f2fa3f1696159554e06bc5803d2a834))


### Build System

* **deps:** bump actions/upload-artifact from 7.0.1 to 7.0.2 in the github-actions group across 1 directory ([#59](https://github.com/tclavelloux/promy-event-bus/issues/59)) ([d9162f0](https://github.com/tclavelloux/promy-event-bus/commit/d9162f0ba25c2dd431f589540ad24e4371829057))
* **deps:** bump github.com/stretchr/testify from 1.8.4 to 1.12.1 ([#41](https://github.com/tclavelloux/promy-event-bus/issues/41)) ([e51c79d](https://github.com/tclavelloux/promy-event-bus/commit/e51c79dc3501973f9d6d61cf4a08ce53773baef1))
* **deps:** bump gitleaks/gitleaks-action from 2.3.9 to 3.0.0 ([#38](https://github.com/tclavelloux/promy-event-bus/issues/38)) ([6b7c323](https://github.com/tclavelloux/promy-event-bus/commit/6b7c32371a37d2c09342cbf1155e19f10c4bb72e))
* **deps:** bump googleapis/release-please-action from 4.4.1 to 5.0.0 ([#44](https://github.com/tclavelloux/promy-event-bus/issues/44)) ([dcb6774](https://github.com/tclavelloux/promy-event-bus/commit/dcb6774394bc4ad0f7e35934216cdb3445a8acba))
* **deps:** bump the gomod group across 1 directory with 6 updates ([#64](https://github.com/tclavelloux/promy-event-bus/issues/64)) ([31227d1](https://github.com/tclavelloux/promy-event-bus/commit/31227d130ce5f0591f12dba55feb7c3ce1f9e9b8))
* **deps:** bump the gomod group with 6 updates ([#56](https://github.com/tclavelloux/promy-event-bus/issues/56)) ([74c02ab](https://github.com/tclavelloux/promy-event-bus/commit/74c02abe71138db60f19bc81312a86e1541a51ef))
* pin go-test-coverage to the version CI runs ([#49](https://github.com/tclavelloux/promy-event-bus/issues/49)) ([ac7645c](https://github.com/tclavelloux/promy-event-bus/commit/ac7645c371ed38a883415948d0b849f466654029))


### Continuous Integration

* add the workflow-sync fleet check and harden registry.yaml ([#53](https://github.com/tclavelloux/promy-event-bus/issues/53)) ([173ad20](https://github.com/tclavelloux/promy-event-bus/commit/173ad200d5013048af8f2f71ab254073ce850b7b))
* automate dependency bumps and vuln fixes ([#55](https://github.com/tclavelloux/promy-event-bus/issues/55)) ([2cf8c7b](https://github.com/tclavelloux/promy-event-bus/commit/2cf8c7bbd5b79e0a6f055d5cc4ab5e31a2359f1c))
* **governance:** adopt hooks/v1.6.0 for release-please and fleet-bot automerge ([#66](https://github.com/tclavelloux/promy-event-bus/issues/66)) ([f786b6f](https://github.com/tclavelloux/promy-event-bus/commit/f786b6fc95749d84c6b6be704c16285c645dd375))
* pin actions to SHAs, add Dependabot and the PR-title check ([#37](https://github.com/tclavelloux/promy-event-bus/issues/37)) ([0462a6b](https://github.com/tclavelloux/promy-event-bus/commit/0462a6b7da3f420d0213d98e3b26a356b4069ae5))
* pin release-please-action by SHA ([#43](https://github.com/tclavelloux/promy-event-bus/issues/43)) ([44f5364](https://github.com/tclavelloux/promy-event-bus/commit/44f5364b22a2303617bcb3788f8228b164ffae09))
* scope permissions, add concurrency and job timeouts ([#36](https://github.com/tclavelloux/promy-event-bus/issues/36)) ([bf4281d](https://github.com/tclavelloux/promy-event-bus/commit/bf4281de391c6f48121a86f2168e12216c834d4a))


### Documentation

* align hook documentation with the fleet ([#47](https://github.com/tclavelloux/promy-event-bus/issues/47)) ([65d2948](https://github.com/tclavelloux/promy-event-bus/commit/65d2948cabf2e4404cf02904f5afd1ee030e4b38))
* **claude:** correct the event contract, retry/DLQ and new-event steps ([#52](https://github.com/tclavelloux/promy-event-bus/issues/52)) ([2643de6](https://github.com/tclavelloux/promy-event-bus/commit/2643de6ca86317a9220763c025aeeaa41156f894))
* correct the Redis DSN and layout in CLAUDE.md, add CI notes, remove the release guide ([#48](https://github.com/tclavelloux/promy-event-bus/issues/48)) ([67d017b](https://github.com/tclavelloux/promy-event-bus/commit/67d017b27382e5ca53e21ecc0d24607c21047a76))
* restructure the README to the canonical layout and refresh it ([#50](https://github.com/tclavelloux/promy-event-bus/issues/50)) ([191e5a9](https://github.com/tclavelloux/promy-event-bus/commit/191e5a9252ecab29ee05a7f8ef0608b94f28b555))


### Miscellaneous

* align pre-commit hook installation with the fleet ([#46](https://github.com/tclavelloux/promy-event-bus/issues/46)) ([1cfb690](https://github.com/tclavelloux/promy-event-bus/commit/1cfb690b3154a1736e107b1f3c94f9f26eb47480))
* **changelog:** remove entries duplicated by merge commits ([#34](https://github.com/tclavelloux/promy-event-bus/issues/34)) ([7cfe73e](https://github.com/tclavelloux/promy-event-bus/commit/7cfe73ecbe92df27eaf85e257690ea65c694bb8d))
* **pre-commit:** bump promy-github-workflows hooks to v1.3.0 ([#63](https://github.com/tclavelloux/promy-event-bus/issues/63)) ([cd83384](https://github.com/tclavelloux/promy-event-bus/commit/cd833848dfed7f6e58c20cb01385bb644d6c39f8))

## [0.4.1](https://github.com/tclavelloux/promy-event-bus/compare/v0.4.0...v0.4.1) (2026-09-12)


### Features

* **docker-compose:** small comment ([d3658b4](https://github.com/tclavelloux/promy-event-bus/commit/d3658b4203d741c214ad4aa1bd77ab760d696c1d))
* **docker:** add RedisInsight for local stream observability ([7755ce4](https://github.com/tclavelloux/promy-event-bus/commit/7755ce4aad401878dd2e12a2ff187c61788ddc9d))


### Bug Fixes

* **deps:** remediate called CVEs, bump go to 1.26.8 ([#33](https://github.com/tclavelloux/promy-event-bus/issues/33)) ([2e812a3](https://github.com/tclavelloux/promy-event-bus/commit/2e812a34b2b26f70a069a9f3fb774d528819f856))
* stop Redis-backed tests from hardcoding localhost:6379 ([#28](https://github.com/tclavelloux/promy-event-bus/issues/28)) ([89d8253](https://github.com/tclavelloux/promy-event-bus/commit/89d82538c396f2d110fd6c7ffbe10acf30ee8fdf)), closes [#25](https://github.com/tclavelloux/promy-event-bus/issues/25)


### Tests

* raise coverage floor to 85% ([#29](https://github.com/tclavelloux/promy-event-bus/issues/29)) ([baad4b4](https://github.com/tclavelloux/promy-event-bus/commit/baad4b4d80df0a4488a92c261edca4b761ba3ed6)), closes [#27](https://github.com/tclavelloux/promy-event-bus/issues/27)


### Continuous Integration

* add lint, test, and gitleaks GitHub Actions workflow ([416ed14](https://github.com/tclavelloux/promy-event-bus/commit/416ed147e2e9c7d2c7e8b23611d90dd528b29c4f))
* **coverage:** adopt go-coverage.yml v2 ([#32](https://github.com/tclavelloux/promy-event-bus/issues/32)) ([2d8e7d2](https://github.com/tclavelloux/promy-event-bus/commit/2d8e7d2439e1863c93df76a1040329273e28eee7))
* **dev:** add CI workflow and pre-commit quality hooks ([d71cde2](https://github.com/tclavelloux/promy-event-bus/commit/d71cde2faede51fc5016860aae67dcc3ab6783c4))
* enforce coverage gate via reusable workflow ([#26](https://github.com/tclavelloux/promy-event-bus/issues/26)) ([b16f81f](https://github.com/tclavelloux/promy-event-bus/commit/b16f81f2df83bf2653b638a92769c25ded57e830))
* **lint:** adopt the reusable go-lint workflow ([#30](https://github.com/tclavelloux/promy-event-bus/issues/30)) ([859f074](https://github.com/tclavelloux/promy-event-bus/commit/859f0742401d7f998b3d4978ce0c1d7641529afe))


### Documentation

* **readme:** note RedisInsight in make up and bump watermark ([87c5dbb](https://github.com/tclavelloux/promy-event-bus/commit/87c5dbbfaa87d57684290dbaa3c74a0c10c21164))


### Miscellaneous

* **dev:** add pre-commit hooks with conventional-commits and gitleaks ([35536b1](https://github.com/tclavelloux/promy-event-bus/commit/35536b14d7554246b4ebdeffaa09feabc572906c))
* **hooks:** add a pinned local lint gate ([#31](https://github.com/tclavelloux/promy-event-bus/issues/31)) ([e16aecb](https://github.com/tclavelloux/promy-event-bus/commit/e16aecbcd98a2e71794871ea9e70075d1e6b9943))

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
* **docs:** remove outdated cursor git workflow docs ([dc45743](https://github.com/tclavelloux/promy-event-bus/commit/dc457430cad2aa7855495e5d1e0d1cf823bb1fbb))
* **gh actions:** wire token ([363ef5b](https://github.com/tclavelloux/promy-event-bus/commit/363ef5b9550abdd908f9631d4710f8817dd25278))
* port infrastructure and tooling adaptations ([f9a12cd](https://github.com/tclavelloux/promy-event-bus/commit/f9a12cdf9aedd01ae6af541112af2de32c1d1540))

## [0.3.1](https://github.com/tclavelloux/promy-event-bus/compare/v0.3.0...v0.3.1) (2025-12-13)


### Features

* **deploy:** add Railway deployment for Redis service ([93d02a6](https://github.com/tclavelloux/promy-event-bus/commit/93d02a6062b52e8368b9debd2d97cdd1e0fb5a9f))
* **events:** update promotion event date handling ([fb954e0](https://github.com/tclavelloux/promy-event-bus/commit/fb954e033ac6dc69cdd76376b984024dba0a0f00))


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


### Miscellaneous

* **main:** release 0.2.0 ([7b3ba83](https://github.com/tclavelloux/promy-event-bus/commit/7b3ba833f12ee397d278a30924d0247b2f3416e7))

## [0.2.0](https://github.com/tclavelloux/promy-event-bus/compare/v0.1.0...v0.2.0) (2025-11-19)


### Features

* **events:** add validation tags to user events ([b54e4af](https://github.com/tclavelloux/promy-event-bus/commit/b54e4af1714676de91e4355535b31e1e9b3818ac))
* **events:** event schemas ([1dcb760](https://github.com/tclavelloux/promy-event-bus/commit/1dcb7609cf666d40ef18ed6e819de46233565699))
* **events:** event schemas for product, promotion and user preferences updates ([bdf5064](https://github.com/tclavelloux/promy-event-bus/commit/bdf50644f5a2364dbc9fe1c6c3f123dba28233e9))
* **redis:** integrate struct validation in publisher ([5475650](https://github.com/tclavelloux/promy-event-bus/commit/547565049c98479e24a1cb26f5e45f0cc45028a5))
