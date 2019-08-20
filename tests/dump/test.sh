#! /bin/bash

# Test the dump restore functionality
#
# Steps taken:
# - Create a chain
# - Create some code and events
# - Dump chain
# - Stop chain and delete
# - Restore chain from dump
# - Check all bits are present (account, namereg, code, events)

set -e
set -x

burrow_dump="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
tmp_dir=$(mktemp -d 2>/dev/null || mktemp -d -t 'tmpdumpXXX')
trap "rm -rf $tmp_dir" EXIT

cd $tmp_dir
rm -rf .burrow genesis.json burrow.toml burrow.log

burrow_bin=${burrow_bin:-burrow}

title="Creating new chain..."
echo -e "\n${title//?/-}\n${title}\n${title//?/-}\n"

$burrow_bin spec -n "Fresh Chain" -r1 | $burrow_bin configure -n BurrowTestDumpNode -e "always" -s- --separate-genesis-doc genesis.json > burrow.toml

$burrow_bin start 2>> burrow.log &

burrow_pid=$!
function kill_burrow {
	kill -KILL $burrow_pid
	rm -rf $tmp_dir
}
trap kill_burrow EXIT

sleep 1

title="Creating code, events and names..."
echo -e "\n${title//?/-}\n${title}\n${title//?/-}\n"

$burrow_bin deploy -o '' -a Root_0 --dir $burrow_dump deploy.yaml

title="Dumping chain..."
echo -e "${title//?/-}\n${title}\n${title//?/-}\n"

$burrow_bin dump remote -b dump.bin
$burrow_bin dump remote dump.json
height=$(head -1  dump.json | jq .Height)

kill $burrow_pid

# Now we have a dump with out stuff in it. Delete everything apart from
# the dump and the keys
mv genesis.json genesis-original.json
rm -rf .burrow burrow.toml

title="Create new chain based of dump with new name..."
echo -e "\n${title//?/-}\n${title}\n${title//?/-}\n"

$burrow_bin configure -m BurrowTestRestoreNode -e "always" -n "Restored Chain" --genesis genesis-original.json --separate-genesis-doc genesis.json --restore-dump dump.json > burrow.toml

$burrow_bin restore dump.json
$burrow_bin start 2>> burrow.log &
burrow_pid=$!
sleep 13

title="Dumping restored chain for comparison..."
echo -e "\n${title//?/-}\n${title}\n${title//?/-}\n"

$burrow_bin dump remote --height $height dump-after-restore.json

kill $burrow_pid

#
# The contract emits an event which contains the hex string DEADCAFE. So,
# this string should be present both in contract code and as an emitted
# event. We should have two in our dump.
#
deadcafe=$(grep DEADCAFE dump.json | wc -l)
if [[ $deadcafe -ne 2 ]]; then
	echo "DUMP FAILURE -- missing DEADCAFE"
	exit 1
fi

if cmp dump.json dump-after-restore.json
then
	title="Done."
	echo -e "\n${title//?/-}\n${title}\n${title//?/-}\n"
else
	echo "RESTORE FAILURE"
	echo "restored dump is different"
	diff -u dump.json dump-after-restore.json
	exit 1
fi
