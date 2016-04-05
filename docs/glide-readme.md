# Using glide with this repository
We use the (glide)[https://github.com/Masterminds/glide] tool to manage our go
dependencies.

In this repo we maintain a set of vendored dependencies under vendor/. Make sure
you have environment variable `GO15VENDOREXPERIMENT` set:

```bash
export GO15VENDOREXPERIMENT=1
```

## Installing dependencies
To (re)install dependencies from scratch based on the values locked in by 
glide.lock run:

```bash
glide install -s -u
```

Where `-s` strips VCS files, in the case of git this stops vendored dependencies
from being treated as submodules that can cause problems, and `-u` updates vendored
dependencies, that install otherwise skips when dealing with a vendor/ without
git roots.

To update dependencies - and store the updated versions in glide.lock you can run:

```
glide up -s -u
```

This will update the versions according to the specification in glide.yaml, which
may: update to the latest available, update up to some version bound, or may keep
exact same version if hooked to a specific commit.

*Beware updating dependencies should be considered destructive* and should only 
be done deliberately and not as part of unrelated updates to the code as a matter
of course.

## Running tests
Running `go test ...` from the root of the repository will try to execute the tests
belonging to all packages under vendor/. Not only is this probably not what you want
but those tests are likely to break because the the test runner may be unable to find
their own nested vendored dependencies.

Instead use:

```bash
glide novendor | xargs go test
```

Where `glide novendro` returns a newline-delimited list of packages in this project
excluding vendor of the form `./<package name>/...`.
