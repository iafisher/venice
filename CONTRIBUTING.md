# Contributing to Venice
Welcome! Thank you for your interest in Venice. There are many ways in which you can contribute to the project, including:

- Reporting a bug
- Requesting a feature
- Sending a pull request

Please see the relevant section of this document for instructions.

Note that any contributions will be released under the terms of Venice's [license](https://github.com/iafisher/venice/blob/master/LICENSE).


## Reporting a bug
Please file bugs using the [GitHub issue tracker](https://github.com/iafisher/venice/issues). You should include (a) the exact Venice code you ran, stripped down to the minimum, (b) what result you got, (c) what result you expected to get, and (d) what platform and version you are using (e.g., Windows 10, macOS Catalina, Ubuntu Linux 20.04).


## Requesting a feature
Features can be requested by filing an issue on the [GitHub issue tracker](https://github.com/iafisher/venice/issues).


## Sending a pull request
Unless your proposed change is trivial (e.g., a typo fix), you should start a discussion *before* you start working on the pull request. Appropriate places to do so are the [GitHub issue tracker](https://github.com/iafisher/venice/issues) and the [venice-users Google group](https://groups.google.com/g/venice-users).

Once you're ready to begin work, the [development guide](https://github.com/iafisher/venice/tree/master/docs/development.md) has everything you need to know about working on the Venice codebase, and you can follow the [GitHub docs](https://docs.github.com/en/github/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request) to create a pull request.

Two things to keep in mind:

- Your pull request must be formatted with `go fmt`. See the "Style" section of the development guide.
- Your pull request must include tests. See the "Testing" section of the development guide.

Some automated checks will be run on your pull request to make sure that it builds and passes the test suite. Then, your pull request will be reviewed by a project owner. Please don't be discouraged if you receive lots of comments - we aim for correct, robust, and clear code, and it can sometimes take several tries to get it right. Ideally, the entire Venice codebase should be of uniform quality and style, no matter how many individual developers contributed to it.
