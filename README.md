# Spider-Man
[![Build Status](https://travis-ci.org/robcharlwood/spider-man.svg?branch=main)](https://travis-ci.org/robcharlwood/spider-man/)

Simple web-crawler CLI written in Go

### Checkout and setup
To work with this codebase, you will require the below to be setup and configured on your machine.

* ``golang`` at version ``1.15.0`` or newer (if you want to make changes to the crawler code)
* ``make`` - If you wish to use the repo's included ``Makefile``

If you wish to use the pre-commit hooks included in this repository, you'll also need:
* ``python`` at version ``3.8.2`` - I suggest installing via ``pyenv`` for isolation.
* Python's ``poetry`` library installed against ``3.8.2``

To set this codebase up on your machine, you can run the following commands:

```bash
git clone git@github.com:robcharlwood/spider-man.git
cd spider-man
```

If you'd like the pre-commit hooks installed, then you also need to run:

```bash
make install
```

The ``Makefile`` checks that you have all the required things at the required versions and ``make install`` will setup a local ``.venv`` environment and install ``pre-commit`` into it.
It will then setup the pre-commit hooks so that the below commands are run locally before a commit can be pushed:

* ``go fmt``
* ``go vet``
* ``go test``
* ``go build``
* ``go mod tidy``

### Tests
The crawler test suite can be run using the below command:

``` bash
go test
```

### Continuous Integration

This project uses [Travis CI](http://travis-ci.org/) for continuous integration. This platform runs the project tests automatically when a PR is raised or merged.

## Versioning

This project uses [git](https://git-scm.com/) for versioning. For the available versions,
see the [tags on this repository](https://github.com/robcharlwood/spider-man/tags).

## Authors

* Rob Charlwood

## Changes

Please see the [CHANGELOG.md](https://github.com/robcharlwood/spider-man/blob/main/CHANGELOG.md) file additions, changes, deletions and fixes between each version

## License

This project is licensed under the CC0-1.0 License - please see the [LICENSE.md](https://github.com/robcharlwood/spider-man/blob/main/LICENSE) file for details
