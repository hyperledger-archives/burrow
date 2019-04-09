## Release checklist

* First of all make sure everyone is happy with doing a release now. 
* Update project/history.go with the latest releases notes and version. Run `make CHANGELOG.md NOTES.md` and make sure this is merged to develop.
* On the develop branch, run `make ready_for_pull_request`. Check for any modified files.
* Using the github.com web interface, create a pull request for master <= develop (so merging latest develop into master)
* Get someone to merge it. They should check that all commits from develop are included using `git log --oneline origin/develop ^origin/master`
* Once master is update to date, switch to master locally run `make tag_release`. This will push the tag which kicks of the release build.
* Optionally send out email on hyperledger burrow mailinglist. Agreements network email should be sent out automatically.