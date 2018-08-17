pragma solidity ^0.4.20;

contract Driver {
    string _name;
    bytes32[] _ownedCars;

    function getName() public view returns (string) {
        return _name;
    }
    function getCars() public view returns (bytes32[]) {
        return _ownedCars;
    }
    function setName(string name) public {
        _name = name;
    }
    function addCar(bytes32 car) public {
        _ownedCars.push(car);
    }

    function addCars(bytes32[] cars) public {
        for (uint index = 0; index < cars.length; index++) {
            _ownedCars.push(cars[index]);
        }
    }
    
    function getCarAmount() public view returns (uint) {
        return _ownedCars.length;
    }
}

