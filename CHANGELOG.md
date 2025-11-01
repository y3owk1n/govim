# Changelog

## [1.0.3](https://github.com/y3owk1n/govim/compare/v1.0.2...v1.0.3) (2025-11-01)


### Bug Fixes

* real electron support (mess) ([#15](https://github.com/y3owk1n/govim/issues/15)) ([6c43bc1](https://github.com/y3owk1n/govim/commit/6c43bc17278f9ba23c3728a3af55b8e7567d5842))

## [1.0.2](https://github.com/y3owk1n/govim/compare/v1.0.1...v1.0.2) (2025-11-01)


### Bug Fixes

* allow to unset global hotkeys ([#10](https://github.com/y3owk1n/govim/issues/10)) ([a0fc0bf](https://github.com/y3owk1n/govim/commit/a0fc0bfec23673e07b9981fdf2b55a2a9c3ce4ab))
* ensure correct config exposes to `status` command ([#8](https://github.com/y3owk1n/govim/issues/8)) ([f2e3f1e](https://github.com/y3owk1n/govim/commit/f2e3f1ead92dc412bddabb0082bd0f851a3eeb4e))
* remove reload config ([#11](https://github.com/y3owk1n/govim/issues/11)) ([a493743](https://github.com/y3owk1n/govim/commit/a4937433777f40aee7ced39653aa168a4f738020))
* remove root run and requires `launch` command to start daemon ([#12](https://github.com/y3owk1n/govim/issues/12)) ([efbf00a](https://github.com/y3owk1n/govim/commit/efbf00a88388cd66618ec8fe79f69ce117728902))

## [1.0.1](https://github.com/y3owk1n/govim/compare/v1.0.0...v1.0.1) (2025-11-01)


### Bug Fixes

* ensure action to run with `cgo` enabled ([#6](https://github.com/y3owk1n/govim/issues/6)) ([0963c66](https://github.com/y3owk1n/govim/commit/0963c66e5ce204048f91c06ffedf20fec572f37f))
* ensure checking from ./bin instead of ./build ([#4](https://github.com/y3owk1n/govim/issues/4)) ([8ee6942](https://github.com/y3owk1n/govim/commit/8ee6942ab960d81a0b3f3409443cbd076162d3e8))

## 1.0.0 (2025-11-01)


### Features

* actually implement middle click ([3783a60](https://github.com/y3owk1n/govim/commit/3783a6008ebfc0aad4718d4bc151f8e2ee58d1cb))
* add ci ([#1](https://github.com/y3owk1n/govim/issues/1)) ([0aab5c0](https://github.com/y3owk1n/govim/commit/0aab5c0e0d23f8c76f8b116e777cdd04fc05f239))
* add hint support for dock and menubar ([7c08d53](https://github.com/y3owk1n/govim/commit/7c08d533304fffa8eafccde51bffd8a31789161e))
* allow additional clickable roles in config ([f893bb4](https://github.com/y3owk1n/govim/commit/f893bb4fe3da2680cf01ea4a369b66e7cfad22f2))
* better hints ([4685a7f](https://github.com/y3owk1n/govim/commit/4685a7f1e219cc73f1673192a915afdaa4142746))
* init project with implementation ([43d255e](https://github.com/y3owk1n/govim/commit/43d255ee3995f4dbaea09ab5c75c5eeb6454a404))
* initial try to support electron ([86db6f7](https://github.com/y3owk1n/govim/commit/86db6f7611b7059fd54827cf87bf8a2e74ed809a))
* more hint modes ([3badbbb](https://github.com/y3owk1n/govim/commit/3badbbb22254ed7a1342ddf1f895eeb21957d1e1))
* nicer action hints ([2a712b0](https://github.com/y3owk1n/govim/commit/2a712b0b71bae4aa7a9d4ef54c799cac0ca0d7df))
* nicer hint with arrow ([e83c5b8](https://github.com/y3owk1n/govim/commit/e83c5b86a72acb9a2008d79fcf6fd9a135170c03))
* only check visible child for ax ([2c1252b](https://github.com/y3owk1n/govim/commit/2c1252b79a7229d81854aaebebe7178f60ab8b27))
* support `-v` and `--version` ([ba29ea1](https://github.com/y3owk1n/govim/commit/ba29ea183b6fbb5ecb48e2d865037378a4afaf35))
* support backspace to go back during hint mode ([8399942](https://github.com/y3owk1n/govim/commit/8399942e996f0d192478de8bc34eb508cfcb4df5))
* support global and per-app ax roles ([677e1c4](https://github.com/y3owk1n/govim/commit/677e1c4e3f4f7bbd63aa2ac9087ed26ba012204c))
* switch to cobra-cli with IPC socket for normal actions ([958989e](https://github.com/y3owk1n/govim/commit/958989e21f325fe30c85d27cd5721d9b6b9df426))


### Bug Fixes

* actually properly support chrome and electron ([5f6ec2d](https://github.com/y3owk1n/govim/commit/5f6ec2d22aa0d3893894507ec6db3be4e64d52b6))
* add fallback and validation for empty hint_characters ([36b8c64](https://github.com/y3owk1n/govim/commit/36b8c645ba5ea29953946201e76d6dda3a864067))
* configurable roles for scroll and hint ([735c98c](https://github.com/y3owk1n/govim/commit/735c98ceffe8a60392467b462ba8b686940f4327))
* ensure config for hint styles are passed through ([8679355](https://github.com/y3owk1n/govim/commit/867935560a17f9d2fbce1978ee581dfda9fddae3))
* ensure event loop runs in main thread ([6826fdb](https://github.com/y3owk1n/govim/commit/6826fdb05e6b9373135562d7f6bdb38e21d35ad7))
* ensure to flush event after clicks ([cb1d5ab](https://github.com/y3owk1n/govim/commit/cb1d5ab5c7cf4cdbf5010c87576755752eb48f09))
* explicit validation for scrollAreaByNumber ([bdbce7f](https://github.com/y3owk1n/govim/commit/bdbce7f4992d728b8ab583ffa840245925b8654b))
* fallback to actual click if press action failed ([e186bb1](https://github.com/y3owk1n/govim/commit/e186bb1dad3d41f81ef0c9352e963ca1be555b03))
* make matched hint color text configurable ([fd0980f](https://github.com/y3owk1n/govim/commit/fd0980f9ddd18de75e5e2fc7e09182731aab683b))
* make sure `ctrl-c` actually kills the program ([a566cf9](https://github.com/y3owk1n/govim/commit/a566cf958db4441399df4a446b45b63b69015245))
* more logs for additional accessibility ([407f992](https://github.com/y3owk1n/govim/commit/407f99273a2c9f4c03455482d2f5aedefe29556f))
* refactor scroll magic numbers to be constants ([521df2a](https://github.com/y3owk1n/govim/commit/521df2acfd3e94ea1007b341773d7bbc8ce86365))
* remove `escape` mapping ([a795e2b](https://github.com/y3owk1n/govim/commit/a795e2bfb2d15fd154578d575761210c35f470db))
* remove action mapping from config ([6974c7d](https://github.com/y3owk1n/govim/commit/6974c7db0ce742e4fafa43ef3e3d9ff002f5d95e))
* remove stupid animations ([97a390c](https://github.com/y3owk1n/govim/commit/97a390c1734f5fa86ee6c9928c5c489bf7ce4247))
* remove unneeded smoothscroll function ([fb77bdc](https://github.com/y3owk1n/govim/commit/fb77bdc4163b6568e9dff8647485d74b3de37d2c))
* remove unused config ([23df329](https://github.com/y3owk1n/govim/commit/23df3293eae53f1f542aa5c26f0abf2a680bd51f))
* some improvement for pre-production ([b0e5a79](https://github.com/y3owk1n/govim/commit/b0e5a79cb3fac447482d50058f165c6492dcaeff))
* sort of working scroll mechanism ([cd4376a](https://github.com/y3owk1n/govim/commit/cd4376a83ffb5f6910d8e68f3fc423d32d098b03))
* support 3 characters hint without clashes ([7120ded](https://github.com/y3owk1n/govim/commit/7120ded044132cfe155665d5cfa70122201dea11))
