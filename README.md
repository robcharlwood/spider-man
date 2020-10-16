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
* ``go build``
* ``go mod tidy``

### Tests
The crawler test suite can be run using the below command:

``` bash
make test
```

The smoke integration tests use golden files for comparison to actual output. To make updates to the golden files,
you can run the below command. This will run the tests and update the golden files during the run. Be careful with this,
you will need to ensure any changes made to the golden files reflect what you are expecting to be in there.

```bash
make test_and_update_golden_fixtures
```

### Building binaries

To build a binary for AMD64 linux, you can run the below command:

```bash
make build_linux
```

To build a binary for Mac OSX Darwin AMD64, you can run the below command:

```bash
make build_macosx
```

### Running the CLI

You can run the spider-man CLI to crawl a website on the internet or a site in development locally. The crawler will ignore
all external links, links for subdomains and links that reference certain resources that we don't usually want to crawl e.g

* ``mailto:``
* ``ftp://``
* ``telnet://``
* ``gopher://``

The below examples assume you have built a binary for the CLI. If you just want to run locally via go, simply
replace ``./spider-man`` prefix with ``go run main.go``.

**Example usage**

The most basic usage is shown below:

```bash
./spider-man crawl https://monzo.com
```

**Parallel**

You can configure the number of requests that are sent to the server in parallel using the ``--parallel`` flag.
Simply pass in an integer representing the number of parallel processes to run. Please be a good neighbour and be
sensible with this if you are crawling live websites. :-)

The default value is 5.

```bash
./spider-man crawl https://monzo.com --parallel 10
```

**Depth**

Sometimes you don't want to crawl the whole website and might just want a subset of "layers". To do this, you can
pass the ``depth`` flag. If this value is not set then the crawler will crawl the passed domain until it runs out of
links to crawl.

The default value is 0 (not set)

```bash
./spider-man crawl https://monzo.com --depth 4
```

**Wait**

This flag can be used to configure the delay between making requests to the server. This can be tweaked and used together with
the ``parallel`` flag to ensure that the website you are crawling is not overrun with requests causing a possible DOS.

The default value is 1 second

```bash
./spider-man crawl https://monzo.com --wait 5s
```

### Continuous Integration

This project uses [Travis CI](http://travis-ci.org/) for continuous integration. This platform runs the project tests automatically when a PR is raised or merged. Travis runs tests against linux and macos against go 1.15.

## Versioning

This project uses [git](https://git-scm.com/) for versioning. For the available versions,
see the [tags on this repository](https://github.com/robcharlwood/spider-man/tags).

## Authors

* Rob Charlwood

## Changes

Please see the [CHANGELOG.md](https://github.com/robcharlwood/spider-man/blob/main/CHANGELOG.md) file additions, changes, deletions and fixes between each version.

## License

This project is licensed under the CC0-1.0 License - please see the [LICENSE.md](https://github.com/robcharlwood/spider-man/blob/main/LICENSE) file for details.
