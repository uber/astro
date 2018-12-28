# astro changelog

## 0.4.2 (UNRELEASED, 2018)

* Add Travis configuration, `make lint` and git precommit hook
* Fix issue with `make test` always returning true even when tests fail
* Fix a race condition that could cause failures due to astro downloading the
  same version of Terraform twice

## 0.4.1 (October 3, 2018)

* Output policy changes in unified diff format (#2)
* Add ability to pass additional arbitrary parameters to terraform at the cli (#3)

## 0.4.0 (September 27, 2018)

* Initial public release #2
