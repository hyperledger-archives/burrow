# Proposals and Voting

Proposals are a way of only executing some transactions if it receives enough votes. This can be 
useful if, for example, there are some solidity contracts which are shared between multiple parties. The proposal
(the transactions) are stored on-chain and other members can verify the proposal before voting on it. Once the
proposal receives enough votes, it is instantly and atomically executed.

## Setup

We want a chain with three participants and a root account for executing proposals. So, creates this with:

```shell
burrow spec -v1 -r1 -p3  | burrow configure -s- -w genesis.json > burrow.toml
```

Note that in the genesis doc there is a ProposalThreshold which is set to 3. This can be modified to suit your
needs. We will leave it at three for now. However if you set this to 1, proposals will execute instantly since
a proposal already has one vote once it is created (the proposer itself).

## Create a Proposal

A proposal is a deployment yaml file, with some minor differences. The transactions which are to be proposed 
should be contained in a proposal job, which should have a name and a description. This proposal jobs type has a 
member called jobs. Store the jobs to be proposed here; each entry should have a source address which should
ideally be a dedicated account for proposal. 

No jobs of type `Assert` or `QueryContract` are allowed in a proposal.

This is our Solidity contract we are proposing:

```solidity
pragma solidity > 0.0.0;

contract random {
	function getInt() public pure returns (int) {
		return 102;
	}
}
```

A standard deploy yaml for this contract would be:

```yaml
jobs:
 - name: deploy_random
   deploy:
     contract: random.sol
```

And it would be deployed like so:

```shell
burrow deploy -a Participant_0 random.yaml 
```

Now we would like this to be a proposal. So, it needs to go into a proposal job and have it's source address
set. The deploy yaml will look like:

```yaml
jobs:
 - name: Propose Deploying contract random
   proposal:
     name: random.sol
     description: I says we should deploy random.sol
     jobs:
      - name: deploy_random
        deploy:
          source: Root_0
          contract: random.sol
```

Now, to create this proposal:

```shell
burrow deploy --proposal-create -a Participant_0 propose-random.yaml 
```

The output should end with:

```shell
log_channel=Info message="Creating Proposal" hash=5029B2B06D42A6339FBD9A97A230F914E3F655143C66B647979ACD05A04C8451
```

# Vote for a Proposal

So Participant_0 created a proposal. Now you are Participant_1, and Participant_0 tells you he's got this proposal
he would like you to vote for. So first of all you want to list the current proposals:

```shell
burrow deploy --list-proposals=PROPOSED
```

```shell
log_channel=Info message=Proposal ProposalHash=5029b2b06d42a6339fbd9a97a230f914e3f655143c66b647979acd05a04c8451 Name=random.sol Description="I says we should deploy random.sol" State=PROPOSED Votes=1
```

Now all we have is a hash. We want to know if this is really the change we are looking for. So, we can verify the proposal using the original deployment yaml and solidity files. You will need the same
solidity compiler version for this to work.

```shell
burrow deploy -a Participant_1 --proposal-verify propose-random.yaml 
```

```shell
log_channel=Info message="Proposal VERIFY SUCCESSFUL" votescount=1
log_channel=Info message=Vote no=0 address=0F73E4EF45EC20BDC7CF5A12EC2F32701C642B9C
```

So the proposal is current, and matches the solidity and deployment files we have. We can now review those changes, and once we're happy with it, we can vote on it using:

```shell
burrow deploy -a Participant_1 --proposal-vote propose-random.yaml 
```

# Ratification and Execution

Once Participant_2 has run:

```shell
burrow deploy -a Participant_2 --proposal-vote propose-random.yaml 
```

The contained transactions are executed. This happens in the same block as where the this vote is registered. 

# Executed and Expired Proposals

```shell
burrow deploy --list-proposals=ALL
```

```shell
log_channel=Info message=Proposal ProposalHash=5029b2b06d42a6339fbd9a97a230f914e3f655143c66b647979acd05a04c8451 Name=random.sol Description="I says we should deploy random.sol" State=EXECUTED Votes=3
```

Executing the transactions increased the sequence number of the Root_0 account. The transactions stored in the proposal
depend on the sequence number being current. If Root_0 executed another transaction before the proposal executed, then
the proposal would have become State=EXPIRED and it cannot not be voted any more.
