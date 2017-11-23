pragma solidity >=0.0.0;

// Interface contract below generated from Burrow with 'make snatives'
/**
* Interface for managing Secure Native authorizations.
* @dev This interface describes the functions exposed by the SNative permissions layer in burrow.
* @dev These functions can be accessed as if this contract were deployed at a special address (0x0a758feb535243577c1a79ae55bed8ca03e226ec).
* @dev This special address is defined as the last 20 bytes of the sha3 hash of the the contract name.
* @dev To instantiate the contract use:
* @dev Permissions permissions = Permissions(address(sha3("Permissions")));
*/
contract Permissions {
    /**
    * @notice Adds a role to an account
    * @param _account account address
    * @param _role role name
    * @return result whether role was added
    */
    function addRole(address _account, bytes32 _role) constant returns (bool result);

    /**
    * @notice Removes a role from an account
    * @param _account account address
    * @param _role role name
    * @return result whether role was removed
    */
    function removeRole(address _account, bytes32 _role) constant returns (bool result);

    /**
    * @notice Indicates whether an account has a role
    * @param _account account address
    * @param _role role name
    * @return result whether account has role
    */
    function hasRole(address _account, bytes32 _role) constant returns (bool result);

    /**
    * @notice Sets the permission flags for an account. Makes them explicitly set (on or off).
    * @param _account account address
    * @param _permission the base permissions flags to set for the account
    * @param _set whether to set or unset the permissions flags at the account level
    * @return result the effective permissions flags on the account after the call
    */
    function setBase(address _account, uint64 _permission, bool _set) constant returns (uint64 result);

    /**
    * @notice Unsets the permissions flags for an account. Causes permissions being unset to fall through to global permissions.
    * @param _account account address
    * @param _permission the permissions flags to unset for the account
    * @return result the effective permissions flags on the account after the call
    */
    function unsetBase(address _account, uint64 _permission) constant returns (uint64 result);

    /**
    * @notice Indicates whether an account has a subset of permissions set
    * @param _account account address
    * @param _permission the permissions flags (mask) to check whether enabled against base permissions for the account
    * @return result whether account has the passed permissions flags set
    */
    function hasBase(address _account, uint64 _permission) constant returns (bool result);

    /**
    * @notice Sets the global (default) permissions flags for the entire chain
    * @param _permission the permissions flags to set
    * @param _set whether to set (or unset) the permissions flags
    * @return result the global permissions flags after the call
    */
    function setGlobal(uint64 _permission, bool _set) constant returns (uint64 result);
}

contract permSNative {
  Permissions perm = Permissions(address(sha3("Permissions")));

  function hasBase(address addr, uint64 permFlag) constant returns (bool) {
    return perm.hasBase(addr, permFlag);
  }

  function setBase(address addr, uint64 permFlag, bool value) constant returns (uint64) {
    return perm.setBase(addr, permFlag, value);
  }

  function unsetBase(address addr, uint64 permFlag) constant returns (uint64) {
    return perm.unsetBase(addr, permFlag);
  }

  // not currently tested
  function setGlobal(uint64 permFlag, bool value) constant returns (int pf) {
    return perm.setGlobal(permFlag, value);
  }

  function hasRole(address addr, bytes32 role) constant returns (bool val) {
    return perm.hasRole(addr, role);
  }

  function addRole(address addr, bytes32 role) constant returns (bool added) {
    return perm.addRole(addr, role);
  }

  function removeRole(address addr, bytes32 role) constant returns (bool removed) {
    return perm.removeRole(addr, role);
  }
}