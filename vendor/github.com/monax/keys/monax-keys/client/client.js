var http = require('http');

var host = 'localhost';
var port = 4767;

function httpOptions(method) {
	return {
		host: host,
		port: port,
		path: '/'+method,
		method: 'POST',
	}
}

mainCallback = function(callback){
	return function(response){
	  var str = '';

	  //another chunk of data has been recieved, so append it to str  
	  response.on('data', function (chunk) {
	    str += chunk;
	  });

	  //the whole response has been recieved, so we just print it out here
	  response.on('end', function () {
		  callback(str)
	  });
	}
}

// the callback should handle `response.on('data')` and `response.on('end')`
function generateKey(name, callback){
	var data = {
		type: "ed25519,ripemd160",
		dir: "",
		auth: "",
		name: name,
	};
	var req = http.request(httpOptions('gen'), callback);
	req.write(JSON.stringify(data));
	req.end();
}

// the callback should handle `response.on('data')` and `response.on('end')`
function signByAddr(addr, bytes, auth, callback){
	var data = {
		type: "ed25519,ripemd160",
		dir: "",
		auth: auth,
		addr: addr,
		hash: bytes,
	};
	var req = http.request(httpOptions('sign'), callback);
	req.write(JSON.stringify(data));
	req.end();
}

function signByName(name, bytes, auth, callback){
	var data = {
		type: "ed25519,ripemd160",
		dir: "",
		auth: auth,
		name: name,
		hash: bytes,
	};
	var req = http.request(httpOptions('sign'), callback);
	req.write(JSON.stringify(data));
	req.end();
}

function verifyByAddr(addr, bytes, sig, callback){
	var data = {
		type: "ed25519,ripemd160",
		dir: "",
		addr: addr,
		hash: bytes,
		sig: sig,
	};
	var req = http.request(httpOptions('verify'), callback);
	req.write(JSON.stringify(data));
	req.end();
}

function verifyByName(name, bytes, sig, callback){
	var data = {
		type: "ed25519,ripemd160",
		dir: "",
		name: name,
		hash: bytes,
		sig: sig,
	};
	var req = http.request(httpOptions('verify'), callback);
	req.write(JSON.stringify(data));
	req.end();
}


module.exports = {
	generateKey: generateKey,
	signByAddr: signByAddr,
	signByName: signByName,
	verifyByAddr: verifyByAddr,
	verifyByName: verifyByName,
	mainCallback: mainCallback,
};
