<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"API Breaking" for breaking CLI commands and REST routes used by end-users.
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.
Ref: https://keepachangelog.com/en/1.0.0/
-->


# Changelog

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
<!-- 
    TODO
--> 

## v1.0.5

- ab6681f feat: add solana pricefeeds
- [#62](https://github.com/NibiruChain/pricefeeder/pull/62) - feat: add stNIBI price discovery

## v1.0.4

- [#46](https://github.com/NibiruChain/pricefeeder/pull/46) - refactor: dynamic version and refactor builds
- [#47](https://github.com/NibiruChain/pricefeeder/pull/47) - Suggestion to allow better precision for exchange rates
- [#55](https://github.com/NibiruChain/pricefeeder/pull/55) - fix(priceprovider-bybit): add special exception for blocked regions in bybit test 
- f2de02b chore(github): Add project automation for https://tinyurl.com/25uty9w5
- [#59](https://github.com/NibiruChain/pricefeeder/pull/59) - fix: reset httpmock before register responder
- [#60](https://github.com/NibiruChain/pricefeeder/pull/60) - chore: lint workflow
- [#61](https://github.com/NibiruChain/pricefeeder/pull/61) - feat(metrics): Add custom prometheus port support and log entry

### Dependencies

- Bump `golang.org/x/net` from 0.17.0 to 0.23.0. ([#48](https://github.com/NibiruChain/pricefeeder/pull/48))

## v1.0.3

- [#45](https://github.com/NibiruChain/pricefeeder/pull/45) - feat(observability): add prometheus metrics and detailed logging 
- [#40](https://github.com/NibiruChain/pricefeeder/pull/40) - avoid parsing if EXCHANGE_SYMBOLS_MAP is not define.

## v1.0.2

- 0408661 fix: remove hard-coded unibi:uusd price

## v1.0.1

- [#25](https://github.com/NibiruChain/pricefeeder/pull/25) - Add gateio source feed
- [#38](https://github.com/NibiruChain/pricefeeder/pull/38) - feat: add Okex datasource
- f2da33a update docs
- 51280e2 refactor symbolsFromPairToSymbolMapping name
- cd64945 use consts in config map
- 16eddcf refactor OKX symbol handling
- 379319d feat: add more data sources
- 4646886 fix: gateio test and symbol handling
- 54b1151 feat: add logging to price sources
- ef632d4 Update README.md
- 88c66a1 fix: coingecko error logging
- 98de6d4 chore: disable coingecko by default
- a067102 add api URLs in comments
- 4e6f6f1 add debug log statements for price updates
- 71003f1 chore: add unibi:uusd default config for GateIO
- [#39](https://github.com/NibiruChain/pricefeeder/pull/39) - feat: add bybit data source