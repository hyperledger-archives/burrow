Hello unfortunate inheritor of the JS libs,

Please forgive me, I did my best. This note is an attempt to explain the structure and decisions that led to burrow.js having the form it does.

First the high level overview. Burrow.js is the javascript interface to connect to, and communicate with a burrow chain. Burrow exposes a GRPC interface which at the lowest level this library wraps. Burrow, if you are unaware, runs an EVM implementation which allows us to use solidity as the smart contract language. This choice however means that in order to be useful we need to format data being passed to the grpc endpoints to be packed according to the ABI of the contract in question. So the overarching goal of this library is to handle:

1. Connecting to burrow
2. Allow the deployment of, or connection to contracts
3. Provide objects which have attached functions facilitating the ability to call contract functions from javascript and hide the intermediate steps
4. Allow listening to event coming from the chain as well as query non contract data from burrow (name registry for example)


## Connecting to Burrow

The index.js file is what is loaded when burrow.js is required. It simply loads up the Burrow file which is responsible for actually creating connection objects to Burrow. It also for convenience exposes the createinstance function which returns the Burrow connection object and the utils file which simply contains some ease of use functions. There has been a push to remove some of these exposed util functions as it is felt that they shouldn’t be relied upon. However they have been inherited from the oldest versions of the js libraries so they have remained for the time being to avoid surprises from removing them.

## lib/Burrow.js

This is the main object for connecting to Burrow is here, along with a helpful creator function which returns it. Note that three things are needed for this, first a url for Burrow which if using the creator function can be either an object with host and port or just a string. An account which will be used by this connection for all calls to burrow that require them (including contract calls) and an options object which is solely used by the contract manager.

During creation first the connection to the various GRPC services are set up, the Service object is a wrapper of the GRPC library that will bind copies of the grpc functions after they have been wrapped to tha same names as before. Its honestly a bit opaque and its main purpose is to make setting up listeners on streaming grpc channels easier. If you want to remove it, go for it, you will need to update the handling of all streaming endpoints in namereg, events etc.
The Burrow connection object exposes all of the low level GRPC functions (in case a user wants direct access to them) but also provides some high level interfaces as well. Namely to events, name registry and most usefully contracts.

There is also the creation of a pipe object which might seem bizarre as it simply wraps other functions from the various services. The pipe is another old piece that has been left solely for ease of a feature that was discussed occasionally but never implemented. Namely that the JS libs we have implemented might be compatible with the main ethereum project. The idea was that if the pipe could serve as an interface from the contract manager to the lower level resources then another pipe could be written that would interface to different low level resources such as those connecting to ethereum. This work stalled though and never progressed. There also did not seem to be a huge need for that functionality.

Events and namereg are both just wrappers of lower level services to ease calling them.



## Contract manager

Contracts are arguably the most important piece of the lib. In order to successfully make a call to an on chain contract a series of steps must be followed. First the ABI of the contract’s function you are calling is needed in order to properly format the arguments to the function in a byte string. Second the formatting must actually take place along with constructing the payload. formatting the data is actually performed by ethereum-js but in order to use it correctly, some pre-formatting of the arguments is required. The payload for the query is sent to burrow either to the “call” or “transact” endpoints of the burrow server. The server will process the payload and then send back the result which will often include return data which must then be processed into return objects and passed to the original caller either in a promise or a callback.

Events are another important part of contracts but luckily much simpler. for events you simply pass a handler which will listen to the event stream of all events matching the pegjs query string. the data is similarly unpacked.

Now for some more detail. All of the contract interfacing code is located in /lib/contracts however the so-called “contract manager” which operates as a contract object constructor is located at /lib/contractmanager.js. There are three files in /lib/contracts/, contract.js, function.js, and event.js. contract.js is the code that constructs the actual contract object that will often be passed around in the user code. 

Contract creation is handled by the contract manager (accessible to the user at burrow.contracts) which has two main functions, .new and .deploy. Contract objects should be thought of as interfaces to an abi type which will have a default contract  address (sometimes) which calls and queries will go to, however as an interface it will also expose the ability to supply an address which will override the default if there is any. The idea is that if you have a bunch of contracts which all have the same abi, there is no reason you should need to create a new object for each contract on chain. This is particularly true for contract which come from contract factories as they will by definition all have the same abi. The difference between deploy and new then is simply what occurs before the contract object is returned. With deploy, bytecode for the contract is provided alongside the abi and a new contract is created on the chain first. A contract object is then created which points to this newly created contract by default. new on the other hand will not make any calls to the burrow and will just create a contract object immediately. This means deploy is asynchronous and new is synchronous. A default address can be provided to new which will then be used by future calls.

You will notice some code here dedicated to “handlers” which are passed as an object with two named fields holding functions call and con. Their purpose is to process the return from either calls or contract creation (CONstructor) before being returned to the end user. Note that these handlers will be run after all function calls to that contract object, they also can be provided as an option to the burrow object creation or directly. If a per function handler is desired it was decided that this should be handled by simply using promise chaining for that particular function rather than attempt a complicated name matching procedure.

contract.js as previously mentioned holds the contract object definition itself. During creation of a contract object, the abi and sometimes byte code is saved followed by dynamically adding the functions and events as callable objects by name. This is done in the addFunctionsToContract and addEventsToContract functions. Taking a look at the addFunctions functions we see that multiple objects are returned from a call to create a solidity function. the displayname and typename are only for the dynamic assignment of the function to the contract object, call, encode and decode are the meat of what will be attached. Call performs all the steps of encoding the data sending it to burrow and decoding the response. However due to the use of proxy contracts (contracts which can call other arbitrary contracts) it was important to be able to access the encoding and decoding functionalities on their own so the payload could be encoded and then sent to the proxy contract as an argument.

These functions are bound to the contract object (so they can all use this. properties of the contract object such as the pipe and address). In addition there are multiple types of calls. for example two main types are calls (transactions) and simulated calls the difference being how burrow handles them. These cases are handled by providing a boolean indicating if the call should be simulated or not. If simulated no changes are made to the chain state. Since these have identical payload structures its easiest to simply provide different values to these functions. by default contract.functionname() will use the function abi’s .constant field to determine if the call should be simulated or not. in addition there are .at() forms which allow overwritingthe default address with a provided one. The object is then added to the contract under the displayname.

A similar process is carried out for attaching events.

We are almost done the hard bits. Just a bit further.

Lets discuss the function.js file because it is probably by far the most messy file to consider. THe first several functions are intended to be “functional” in the sense that they take in data and return data without changing the original data. We have functions for constructing function names and signatures putting together the payload for burrow, and encoding and decoding the arguments to be passed as byte data. most of these should be self explanatory. they each serve a role without touching the pathway to burrow. the convert module is used to side step formatting issue that have been the bane of my existence since taking this over. my abiToBurrow and back functions are an attempt to be aware of what type of data it is supposed to be in order to correctly format it before passing it to burrow or ethereum-js. The endless debate has been the 0x prepended to byte and address strings. The short story is that ethereum-js expects hex strings to be 0x prepended or it will attempt asciitohex conversion which causes nightmares but manually prepending 0x to all the hex strings has been unpopular. so the convert module smooths over that. if it is a byte like argument convert will prepend with 0x. It assume you are sending proper hex and will not asciitohex convert for you. suck it up. this has proven the least controversial option.

Anyways the Solidity function is the meat of the function.js. looking through it you will notice that its not the function that gets attached but rather a function which returns functions to get attached to the contract. encode and decode essentially route to the functional versions.  and the call function does the full pathway.

arguments are shifted into variables in order to handle the optional components. Then a promise is constructed which holds the whole pathway to burrow and back. inside the promise decisions are made on if this is a simulated call etc, a post processing callback function is created which will hand the decoding of returns along with processing of execution errors. then the payload is constructed, and the pipe of the contract this function is bound to will be called. The solidity function will either return the promise of it will resolve the promise into a callback is provided. this is a common pattern throughout.

events is basically the same. so yeah.

I think that covers most of it. Once again I’m sorry. I did my best

88224646AB

Long live the DOUG!

Dennis

