### Fixed
- [State] Since State hash is not unique (i.e if we make no writes) by storing the CommitID by AppHash we can overwrite an older CommitID with a newer one leading us to load the wrong tree version to overwrite in case of loading from a checkpoint.

