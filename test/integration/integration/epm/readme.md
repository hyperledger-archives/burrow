## Integration tests for Hyperledger Burrow

Historical much of the high quality test matter were integration tests held in the monax
(formerly eris-cli) tool repository. This tests take the form of solidity contracts and 
files containing automated deployment steps, contract setup, and testing assertions; epm.yaml
files belonging to what was originally called 'Eris Package Manager'. We still refer to them as
'EPM' tests. EPM scripts (epm.yaml files) can be executed by 'monax pkgs do'.

The tests here are taken from the monax tool with a view of bringing them 'in house' inside
Burrow where the core functionality they test resides. For the moment they still depend on
the monax tool but at some point Burrow ought to have its own mechanism for running these
or equivalent tests without an external tool (this could be using Web3 tooling, porting the
tests, or developing another tool that can process the epm.yaml files and deploy contracts)