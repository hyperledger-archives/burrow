var client = require('../');


// generate a key
client.generateKey("bob", client.mainCallback(function(str){
	var resp = JSON.parse(str)
	console.log("address:", resp.Response)
	if (resp.Error){
		// TODO: DO SOMETHING
		console.log(resp.Error);
	}
	var addr = resp.Response;
	// this is what we sign
	var bytes = "0123456789";
	client.signByAddr(addr, bytes, "", mainCallback(function(str){
		var resp = JSON.parse(str)
		if (resp.Error){
			// TODO: DO SOMETHING
			console.log(resp.Error);
		}
		var sig = resp.Response
		client.verifyByAddr(addr, bytes, sig, mainCallback(function(str){
			var resp = JSON.parse(str);
			if (resp.Error){
				// TODO: DO SOMETHING
				console.log(resp.Error);
			}
			if (resp.Response == "true"){
				console.log("PASS");
			} else {
				console.log("FAIL");
				// TODO: exit status 1 ?
			}
		}));

		client.verifyByAddr(addr, "0987654321", sig, mainCallback(function(str){
			var resp = JSON.parse(str);
			if (resp.Error){
				// TODO: DO SOMETHING
				console.log(resp.Error);
			}
			if (resp.Response == "true"){
				console.log("FAIL");
			} else {
				console.log("PASS");
				// TODO: exit status 1 ?
			}
		}));
	}));
}));
