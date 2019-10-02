# astro changelog

## 0.5.0 (UNRELEASED, 2019)

* Add `version` command (#12)
* Adopt options pattern for `astro.NewProject` constructor (#26)
* Refactor and improve integration tests to invoke them directly using cli
  rather than `os.exec` (#26)
* Add Travis configuration, `make lint` and git precommit hook
* Fix `--help` displaying "pflag: help requested" (#1)
* Fix issue with make not recompiling when source files changed
* Fix issue with `make test` always returning true even when tests fail
* Fix race condition that could cause failures due to astro downloading the
  same version of Terraform twice
* Fix module execution errors being printed to the console twice
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

* API: The signature of `astro.NewProject` has changed to now accept a list of
  functional options. This allows us to add new options in the future without
  making a breaking change.

  `astro.NewProject(conf)` should be changed to:
  `astro.NewProject(astro.WithConfig(conf))`

## 0.4.1 (October 3, 2018)

* Output policy changes in unified diff format (#2)
* Add ability to pass additional arbitrary parameters to terraform at the cli (#3)

## 0.4.0 (September 27, 2018)

* Initial public release #2
