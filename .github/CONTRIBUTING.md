# Contributing to `burrow`:
Forked from Docker's [contributing guidelines](https://github.com/docker/docker/blob/master/CONTRIBUTING.md)

## Developer Certificate of Origin

Please always include "signed-off-by" in your commit message and note this constitutes a developer certification you are contributing a patch under Apache 2.0.  Please find a verbatim copy of the Developer Certificate of Origin in this repository [here](.github/DEVELOPER_CERTIFICATE_OF_ORIGIN.md) or on [developercertificate.org](https://developercertificate.org/).

## Bug Reporting

A great way to contribute to the project is to send a detailed report when you encounter an issue. We always appreciate a well-written, thorough bug report, and will thank you for it!

Check that the issue doesn't already exist before submitting an issue. If you find a match, you can use the "subscribe" button to get notified on updates. Add a :+1: if you've also encountered this issue. If you have ways to reproduce the issue or have additional information that may help resolving the issue, please leave a comment.

Also include the steps required to reproduce the problem if possible and applicable. This information will help us review and fix your issue faster. When sending lengthy log-files, post them as a gist (https://gist.github.com). Don't forget to remove sensitive data from your log files before posting (you can replace those parts with "REDACTED").

Our [ISSUE_TEMPLATE.md](ISSUE_TEMPLATE.md) will autopopulate the new issue.

## Contribution Tips and Guidelines

### Pull requests are always welcome (to `develop` rather than `master`).

Not sure if that typo is worth a pull request? Found a bug and know how to fix it? Do it! We will appreciate it. Any significant improvement should be documented as a GitHub issue or discussed in [The Marmot Den](https://slack.monax.io) Slack community prior to beginning.

We are always thrilled to receive pull requests (and bug reports!) and we do our best to process them quickly. 

## Conventions

Fork the repository and make changes on your fork in a feature branch (branched from develop), create an issue outlining your feature or a bug, or use an open one.

    If it's a bug fix branch, name it something-XXXX where XXXX is the number of the issue.
    If it's a feature branch, create an enhancement issue to announce your intentions, and name it something-XXXX where XXXX is the number of the issue.

Submit unit tests for your changes. Go has a great test framework built in; use it! Take a look at existing tests for inspiration. Run the full test suite on your branch before submitting a pull request.

Update the documentation when creating or modifying features. Test your documentation changes for clarity, concision, and correctness, as well as a clean documentation build. 

Write clean code. Universally formatted code promotes ease of writing, reading, and maintenance. Always run `gofmt -s -w file.go` on each changed file before committing your changes. Most editors have plug-ins that do this automatically.

Pull request descriptions should be as clear as possible and include a reference to all the issues that they address.

Commit messages must start with a short summary (max. 50 chars) written in the imperative, followed by an optional, more detailed explanatory text which is separated from the summary by an empty line.

Code review comments may be added to your pull request. Discuss, then make the suggested modifications and push additional commits to your feature branch. 

Pull requests must be cleanly rebased on top of develop without multiple branches mixed into the PR.

*Git tip:* If your PR no longer merges cleanly, use `git rebase develop` in your feature branch to update your pull request rather than merge develop.

Before you make a pull request, squash your commits into logical units of work using `git rebase -i` and `git push -f`. A logical unit of work is a consistent set of patches that should be reviewed together: for example, upgrading the version of a vendored dependency and taking advantage of its now available new feature constitute two separate units of work. Implementing a new function and calling it in another file constitute a single logical unit of work. The very high majority of submissions should have a single commit, so if in doubt: squash down to one.

After every commit, make sure the test suite passes. Include documentation changes in the same pull request so that a revert would remove all traces of the feature or fix.

### Merge approval

We use LGTM (Looks Good To Me) in commands on the code review to indicate acceptance. 

## Errors and Log Messages Style

TODO

## Coding Style

Unless explicitly stated, we follow all coding guidelines from the Go community. While some of these standards may seem arbitrary, they somehow seem to result in a solid, consistent codebase.

It is possible that the code base does not currently comply with these guidelines. We are not looking for a massive PR that fixes this, since that goes against the spirit of the guidelines. All new contributions should make a best effort to clean up and make the code base better than they left it. Obviously, apply your best judgement. Remember, the goal here is to make the code base easier for humans to navigate and understand. Always keep that in mind when nudging others to comply.

* All code should be formatted with `gofmt -s`.
* All code should follow the guidelines covered in [Effective Go](https://golang.org/doc/effective_go.html) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
* Comment the code. Tell us the why, the history and the context.
* Document all declarations and methods, even private ones. Declare expectations, caveats and anything else that may be important. If a type gets exported, having the comments already there will ensure it's ready.
* Variable name length should be proportional to it's context and no longer. noCommaALongVariableNameLikeThisIsNotMoreClearWhenASimpleCommentWouldDo. In practice, short methods will have short variable names and globals will have longer names.
* No underscores in package names. If you need a compound name, step back, and re-examine why you need a compound name. If you still think you need a compound name, lose the underscore.
* No utils or helpers packages. If a function is not general enough to warrant its own package, it has not been written generally enough to be a part of a `util` package. Just leave it unexported and well-documented.
* All tests should run with `go test` and outside tooling should not be required. No, we don't need another unit testing framework. Assertion packages are acceptable if they provide real incremental value.
