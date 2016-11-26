# Contributing to `eris-cli`:
Forked from Docker's [contributing guidelines](https://github.com/docker/docker/blob/master/CONTRIBUTING.md)

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

The below guidelines are more observations of how things are done now rather than strict rules to follow. These are collections of facts to make your code look better and better fit the project. Again, as with coding guidelines, please apply your best judgement.

#### Errors

* Error messages should be short and concise, not containing the `\n`, `\t`, `=>`, or other formatting characters, and not ending with a period or a colon. 
* Ideally, it should fit one line and be less than 80 characters long. 
* If you really need to make it multi-line, use back tick quotes or Go text templates.
* Prefer the present tense to the past tense or the subjunctive mood.
* Multiple sentences within the same error are separated with a dot and a space.

  ```
  fmt.Errorf("I don't know that service. Please retry with a known service")
  ```
* Returned error messages from top level functions (the ones invoked via [Cobra](https://github.com/spf13/cobra/cobra) package and from the `cmd` subdirectory) should start with a capital letter, state the nature of the problem, and, if necessary, include the lower level error (separated from the main message via a colon and a space). The message should be stated from the point of view of the software user and don't include names of functions, packages, or terms which are not found somewhere in the tutorials:

  ```
  return fmt.Errorf("I cannot find that service. Please check the service name you sent me")
  return fmt.Errorf("Could not add ssh.exe to PATH: %v", err)
  ```
* Returned error messages from package level or utility functions which in turn be used by the top level functions should start with a small letter and, if necessary, include the lower level error (separated from the main message via a colon and a space) or use prefabricated errors. The message should be stated from the point of view of the package or library user (Eris developer):

  ```
  return fmt.Errorf("there is no chain checked out")
  return fmt.Errorf("cannot migrate directories: %v", err)
  
  var ErrNoSuchVolume = errors.New("no such volume")
  return ErrNoSuchVolume
  ```
* Returned errors from the Docker Client library ([github.com/fsouza/go-dockerclient](https://github.com/fsouza/go-dockerclient)) should wrap that error with the `util.DockerError` function to remove the word "API" and HTTP status code from the error message:

  ```
  return util.DockerError(err)
  return DockerError(DockerClient.StartContainer(name, nil))
  ```

* Errors that are printed to the console (that is not returned) should use the `logrus.Error` function.

#### Log Messages

* Log messages should be complete sentences (not standalone nouns or names), not containing the `\n`, `\t`, or other formatting characters, and not ending with a period or a colon. 
* Ideally, the message should fit one line and be less than 80 characters long.
* Use `Info` log level for optional messages software users somehow might benefit from (`--verbose` flag).
* Use `Debug` log level for optional messages targeted at developers only (`--debug` flag).
* Prefer dropping articles from log messages (magazine heading style) to make them shorter
  
  ```
  log.Debug("Getting connection details from environment")
  ```
* Multiple sentences on the same line are separated with a dot and a space:

   ```
   log.Info("Chain not currently running. Skipping")
   ```
* Names of functions should not be present in the log message, but the log message should be unique within the code base or at least easily distinguishable (greppable) for easier debugging
* If you need to log useful data along with the message, use tags (the `WithField` thing):

   ```
   log.WithFields(log.Fields{
      "from": do.Name,
      "to":   do.NewName,
   }).Info("Renaming action")

   log.WithFields(log.Fields{
      "=>":        servName,
      "existing#": containerExist,
      "running#":  containerRun,
   }).Info("Checking number of containers for")
   ```
Both tag name and its description should preferably be lowercase (multiple words separated by spaces). 
* If you need to log just a noun without any statement, use an empty log message:

   ```
   log.WithField("drop", dropped).Debug()
   ```
* The `=>` tag prefix is (generally) for names (containers, chains, servers, data, etc.), which never have the tag `name`.
* `logrus` package always ends a log message with an `\n` character, so there's no need to invoke `log.Infoln` or `log.Infof("\n")` or similar.

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
