pragma solidity >=0.4.24;

contract permSNative {
  Permissions perm = Permissions(address(uint256(keccak256("Permissions"))));

  function hasBase(address addr, uint64 permFlag) public returns (bool) {
    return perm.hasBase(addr, permFlag);
  }

  function setBase(address addr, uint64 permFlag, bool value) public returns (uint64) {
    return perm.setBase(addr, permFlag, value);
  }

  function unsetBase(address addr, uint64 permFlag) public returns (uint64) {
    return perm.unsetBase(addr, permFlag);
  }

  // not currently tested
  function setGlobal(uint64 permFlag, bool value) public returns (int pf) {
    return perm.setGlobal(permFlag, value);
  }

  function hasRole(address addr, string memory role) public returns (bool val) {
    return perm.hasRole(addr, role);
  }

  function addRole(address addr, string memory role) public returns (bool added) {
    return perm.addRole(addr, role);
  }

  function removeRole(address addr, string memory role) public returns (bool removed) {
    return perm.removeRole(addr, role);
  }
}

/**
* Interface for managing Secure Native authorizations.
* @dev This interface describes the functions exposed by the native permissions layer in burrow.
* @dev These functions can be accessed as if this contract were deployed at a special address (0x0A758FEB535243577C1A79AE55BED8CA03E226EC).
* @dev This special address is defined as the last 20 bytes of the sha3 hash of the the contract name.
* @dev To instantiate the contract use:
* @dev Permissions permissions = Permissions(address(uint256(keccak256("Permissions"))));
*/
interface Permissions {
    /**
    * @notice Adds a role to an account
    * @param _account account address
    * @param _role role name
    * @return _result whether role was added
    */
    function addRole(address _account, string calldata _role) external returns (bool _result);

    /**
    * @notice Removes a role from an account
    * @param _account account address
    * @param _role role name
    * @return _result whether role was removed
    */
    function removeRole(address _account, string calldata _role) external returns (bool _result);

    /**
    * @notice Indicates whether an account has a role
    * @param _account account address
    * @param _role role name
    * @return _result whether account has role
    */
    function hasRole(address _account, string calldata _role) external returns (bool _result);

    /**
    * @notice Sets the permission flags for an account. Makes them explicitly set (on or off).
    * @param _account account address
    * @param _permission the base permissions flags to set for the account
    * @param _set whether to set or unset the permissions flags at the account level
    * @return _result is the permission flag that was set as uint64
    */
    function setBase(address _account, uint64 _permission, bool _set) external returns (uint64 _result);

    /**
    * @notice Unsets the permissions flags for an account. Causes permissions being unset to fall through to global permissions.
    * @param _account account address
    * @param _permission the permissions flags to unset for the account
    * @return _result is the permission flag that was unset as uint64
    */
    function unsetBase(address _account, uint64 _permission) external returns (uint64 _result);

    /**
    * @notice Indicates whether an account has a subset of permissions set
    * @param _account account address
    * @param _permission the permissions flags (mask) to check whether enabled against base permissions for the account
    * @return _result is whether account has the passed permissions flags set
    */
    function hasBase(address _account, uint64 _permission) external returns (bool _result);

    /**
    * @notice Sets the global (default) permissions flags for the entire chain
    * @param _permission the permissions flags to set
    * @param _set whether to set (or unset) the permissions flags
    * @return _result is the permission flag that was set as uint64
    */
    function setGlobal(uint64 _permission, bool _set) external returns (uint64 _result);
}
