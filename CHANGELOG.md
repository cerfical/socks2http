# Changelog

## [1.2.0](https://github.com/cerfical/socks2http/compare/v1.1.0...v1.2.0) (2025-07-13)


### Features

* **config:** add shorthands for common flags ([ef7e5e0](https://github.com/cerfical/socks2http/commit/ef7e5e032a2faf71b2de9a8aa67983ea7d3f3bc3))
* **config:** add simplified URL parsing ([d75b0ed](https://github.com/cerfical/socks2http/commit/d75b0edf7dd66bf3bb6e6eb45eaee9d6dd9da359))
* **config:** remove `direct` protocol scheme ([1afa238](https://github.com/cerfical/socks2http/commit/1afa238376476639c16becc55adc533727cced99))
* **config:** replace `proxy-proto` and `proxy-addr` options with `proxy` ([ea696b1](https://github.com/cerfical/socks2http/commit/ea696b1497fd530036f8c111973ed8cbe3ee2b43))
* **config:** replace `server-proto` and `server-addr` options with `server` ([0220b1e](https://github.com/cerfical/socks2http/commit/0220b1e1b14190d5df324bd7ad63d99293fcc64a))
* **proxy:** improve `SOCKS` implementation ([8c2ec07](https://github.com/cerfical/socks2http/commit/8c2ec07597f051cbf2bf85b7799c3a2eb515836c))
* **proxy:** improve proxy client implementation ([51a69a7](https://github.com/cerfical/socks2http/commit/51a69a7e3cae1ea95840e514ecb10dea6fe2563a))
* **proxy:** improve proxy server implementation ([9052fe3](https://github.com/cerfical/socks2http/commit/9052fe39201a24e69d4d045346e91f06337ecf54))


### Bug Fixes

* fix incorrect error messages for SOCKS4 proxy clients ([63805f1](https://github.com/cerfical/socks2http/commit/63805f124d3d766d036ba56b6baae88f049b00ae))
* fix invalid default values for configured proxy routes ([f020908](https://github.com/cerfical/socks2http/commit/f0209089152f55dec447bb1d824dfc1669bb4b82))

## [1.1.0](https://github.com/cerfical/socks2http/compare/v1.0.0...v1.1.0) (2025-06-06)


### Features

* **client:** add basic host-based proxy routing ([11b9a05](https://github.com/cerfical/socks2http/commit/11b9a05c1027d240f3714ea70fe503a1a7e6a030))
* **client:** make proxy routes match loosely ([a78e33e](https://github.com/cerfical/socks2http/commit/a78e33ec93feaa2dfae27175f9a25f28bf080b41))
* **config:** add `--config-file` flag ([d9e19ce](https://github.com/cerfical/socks2http/commit/d9e19ce8306e7666dbb39e039aa86fcc22d39ded))
* **config:** add proxy routes ([734ee2a](https://github.com/cerfical/socks2http/commit/734ee2affa3ef58452937ae90cf522ede5cc3277))

## 1.0.0 (2025-05-31)

### Features

- **client:** add `direct` protocol scheme
- **client:** add `http` protocol scheme
- **client:** add `socks4` protocol scheme
- **client:** add `socks4a` protocol scheme
- **client:** add `socks5` protocol scheme
- **client:** add `socks5h` protocol scheme
- **config:** add `log-level` flag
- **config:** add `proxy-addr` flag
- **config:** add `proxy-proto` flag
- **config:** add `server-addr` flag
- **config:** add `server-proto` flag
- **config:** add `timeout` flag
- **proxy:** add HTTP support
- **proxy:** add SOCKS4 support
- **proxy:** add SOCKS5 support
- **server:** add `http` protocol scheme
- **server:** add `socks` protocol scheme
- **server:** add `socks4` protocol scheme
- **server:** add `socks5` protocol scheme
