'use strict'

function Namereg (burrow) {
  this.burrow = burrow
}

Namereg.prototype.set = function (name, data, lease, callback) {
  var payload = {}
  payload.Input = {Address: Buffer.from(this.burrow.account, 'hex'), Amount: 50000}
  payload.Name = name
  payload.Data = data
  payload.Fee = 5000// 1 * lease * (data.length + 32);
  return this.burrow.transact.NameTxSync(payload, callback)
}

Namereg.prototype.get = function (name, callback) {
  var payload = {Name: name}
  return this.burrow.query.GetName(payload, callback)
}

module.exports = Namereg
