#!/usr/bin/env bash

# -------------------------------------------------------------------
# Set vars (change if used in another repo)

base_name=eris-db
user_name=eris-ltd
docs_site=monax.io
docs_name=./docs/documentation
slim_name=db

# -------------------------------------------------------------------
# Set vars (usually shouldn't be changed)

if [ "$CIRCLE_BRANCH" ]
then
  repo=`pwd`
else
  repo=$GOPATH/src/github.com/$user_name/$base_name
fi
release_min=$(cat $repo/version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')
start=`pwd`

# -------------------------------------------------------------------
# Build

cd $repo
rm -rf $docs_name
go run ./docs/generator.go

if [[ "$1" == "master" ]]
then
  mkdir -p $docs_name/$slim_name/latest
  rsync -av $docs_name/$slim_name/$release_min/ $docs_name/$slim_name/latest/
  find $docs_name/latest -type f -name "*.md" -exec sed -i "s/$release_min/latest/g" {} +
fi

tmp_dir=`mktemp -d 2>/dev/null || mktemp -d -t 'tmp_dir'`
git clone git@github.com:$user_name/$docs_site.git $tmp_dir/$docs_site

rsync -av $docs_name $tmp_dir/$docs_site/content/docs/

# ------------------------------------------------------------------
# Commit and push if there are changes

cd $tmp_dir/$docs_site
if [ -z "$(git status --porcelain)" ]; then
  echo "All Good!"
else
  git add -A :/ &&
  git commit -m "$base_name build number $CIRCLE_BUILD_NUM doc generation" &&
  git push origin master
fi

# ------------------------------------------------------------------
# Cleanup

rm -rf $tmp_dir
cd $start