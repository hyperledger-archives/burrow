
pragma solidity >=0.4.0;

/**
* Interface for managing Secure Native authorizations.
* @dev This interface describes the functions exposed by the SNative permissions layer in burrow.
* @dev These functions can be accessed as if this contract were deployed at a special address (0x0A758FEB535243577C1A79AE55BED8CA03E226EC).
* @dev This special address is defined as the last 20 bytes of the sha3 hash of the the contract name.
* @dev To instantiate the contract use:
* @dev Permissions permissions = Permissions(address(keccak256("Permissions")));
*/
contract Permissions {
    /**
    * @notice Adds a role to an account
    * @param Account account address
    * @param Role role name
    * @return result whether role was added
    */
    function addRole(address Account, string Role) public constant returns (bool Result);

    /**
    * @notice Removes a role from an account
    * @param Account account address
    * @param Role role name
    * @return result whether role was removed
    */
    function removeRole(address Account, string Role) public constant returns (bool Result);

    /**
    * @notice Indicates whether an account has a role
    * @param Account account address
    * @param Role role name
    * @return result whether account has role
    */
    function hasRole(address Account, string Role) public constant returns (bool Result);

    /**
    * @notice Sets the permission flags for an account. Makes them explicitly set (on or off).
    * @param Account account address
    * @param Permission the base permissions flags to set for the account
    * @param Set whether to set or unset the permissions flags at the account level
    * @return result the effective permissions flags on the account after the call
    */
    function setBase(address Account, uint64 Permission, bool Set) public constant returns (uint64 Result);

    /**
    * @notice Unsets the permissions flags for an account. Causes permissions being unset to fall through to global permissions.
    * @param Account account address
    * @param Permission the permissions flags to unset for the account
    * @return result the effective permissions flags on the account after the call
    */
    function unsetBase(address Account, uint64 Permission) public constant returns (uint64 Result);

    /**
    * @notice Indicates whether an account has a subset of permissions set
    * @param Account account address
    * @param Permission the permissions flags (mask) to check whether enabled against base permissions for the account
    * @return result whether account has the passed permissions flags set
    */
    function hasBase(address Account, uint64 Permission) public constant returns (bool Result);

    /**
    * @notice Sets the global (default) permissions flags for the entire chain
    * @param Permission the permissions flags to set
    * @param Set whether to set (or unset) the permissions flags
    * @return result the global permissions flags after the call
    */
    function setGlobal(uint64 Permission, bool Set) public constant returns (uint64 Result);
}

contract permSNative {
  Permissions perm = Permissions(address(keccak256("Permissions")));

  function hasBase(address addr, uint64 permFlag) public constant returns (bool) {
    return perm.hasBase(addr, permFlag);
  }

  function setBase(address addr, uint64 permFlag, bool value) public constant returns (uint64) {
    return perm.setBase(addr, permFlag, value);
  }

  function unsetBase(address addr, uint64 permFlag) public constant returns (uint64) {
    return perm.unsetBase(addr, permFlag);
  }

  // not currently tested
  function setGlobal(uint64 permFlag, bool value) public constant returns (int pf) {
    return perm.setGlobal(permFlag, value);
  }

  function hasRole(address addr, string role) public constant returns (bool val) {
    return perm.hasRole(addr, role);
  }

  function addRole(address addr, string role) public constant returns (bool added) {
    return perm.addRole(addr, role);
  }

  function removeRole(address addr, string role) public constant returns (bool removed) {
    return perm.removeRole(addr, role);
  }
}
