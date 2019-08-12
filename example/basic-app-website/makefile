#
# basic-app is intended to provide an example end-to-end use of burrow and burrow.js in a node.js app
#


# One of: Darwin_i386, Darwin_x86_64, Linux_i386, Linux_x86_64
BURROW_ARCH := Linux_x86_64
BURROW_VERSION := 0.21.0
BURROW_RELEASE_URL := "https://github.com/hyperledger/burrow/releases/download/v${BURROW_VERSION}/burrow_${BURROW_VERSION}_${BURROW_ARCH}.tar.gz "
# Set to 'burrow' to use whatever is on PATH instead
BURROW_BIN := bin/burrow

#
# Running the chain
#
# Make a simple single node chain
.PHONY: chain
chain: bin/burrow burrow.toml

# Get the burrow binary
bin/burrow:
	mkdir -p bin
	curl -L ${BURROW_RELEASE_URL} | tar zx -C bin burrow

# Generate the chain
burrow.toml genesis.json:
	${BURROW_BIN} spec --full-accounts 1 | ${BURROW_BIN} configure --genesis-spec=- --separate-genesis-doc=genesis.json > burrow.toml

# Dump account information to file for app
account.json: genesis.json
	jq  '.Accounts[] | select(.Name == "Full_0")' genesis.json > account.json

# Reset burrow state
.PHONY: reset_chain
reset_chain:
	rm -rf .burrow

# Remove burrow chain completely
.PHONY: remove_chain
remove_chain:
	rm -rf burrow.toml genesis.json .keys .burrow

# remake and reset chain
.PHONY: rechain
rechain: | remove_chain chain

.PHONY: start_chain
start_chain: chain
	${BURROW_BIN} start -v0

.PHONY: restart
restart: | rechain start_chain

#
# Deploying the contract
#
deploy.output.json: simplestorage.sol deploy.yaml account.json
	${BURROW_BIN} deploy --address $(shell jq '.Address' account.json)

.PHONY: delete_deploy
delete_deploy:
	rm -rf deploy.output.json

.PHONY: deploy
deploy: deploy.output.json

.PHONY: redeploy
redeploy: | delete_deploy deploy.output.json

#
# Running the app
#

.PHONY: npm_install
npm_install:
	npm install

.PHONY: start_app
start_app: npm_install deploy
	node app.js