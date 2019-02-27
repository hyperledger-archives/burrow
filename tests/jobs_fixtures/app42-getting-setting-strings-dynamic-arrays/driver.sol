pragma solidity >=0.0.0;
contract Driver {
    string _name;
    bytes32[] _ownedCars;

    function getName() public view returns (string memory) {
        return _name;
    }
    function getCars() public view returns (bytes32[] memory) {
        return _ownedCars;
    }
    function setName(string memory name) public {
        _name = name;
    }
    function addCar(bytes32 car) public {
        _ownedCars.push(car);
    }

    function addCars(bytes32[] memory cars) public {
        for (uint index = 0; index < cars.length; index++) {
            _ownedCars.push(cars[index]);
        }
    }
    
    function getCarAmount() public view returns (uint) {
        return _ownedCars.length;
    }
}

