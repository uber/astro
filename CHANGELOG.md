# astro changelog

## 0.5.0 (UNRELEASED, 2018)

* Add Travis configuration, `make lint` and git precommit hook
* Fix `--help` displaying "pflag: help requested" (#1)
* Fix issue with make not recompiling when source files changed
* Fix issue with `make test` always returning true even when tests fail
* Fix race condition that could cause failures due to astro downloading the
  same version of Terraform twice
* Remove godep and move to Go modules (vgo)
* Change configuration syntax for remapping CLI flags to Terraform module
  variables

**Breaking changes:**

* Before, there was a `flag:` option underneath module variables in the project
  configuration that allowed you to modify the name of the flag on the CLI that
  would represent that variable (e.g.: "--environment" could be remapped to
  "--env").

  This has been removed and there is now a completely new section in the YAML
  called "flags". See the "Remapping flags" section of the README for more
  information.

## 0.4.1 (October 3, 2018)

* Output policy changes in unified diff format (#2)
* Add ability to pass additional arbitrary parameters to terraform at the cli (#3)

## 0.4.0 (September 27, 2018)

* Initial public release #2
