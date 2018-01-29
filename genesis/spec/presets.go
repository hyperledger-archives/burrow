package spec

import (
	"fmt"

	"sort"

	"github.com/hyperledger/burrow/permission"
)

// Files here can be used as starting points for building various 'chain types' but are otherwise
// a fairly unprincipled collection of GenesisSpecs that we find useful in testing and development

func FullAccount(index int) GenesisSpec {
	// Inheriting from the arbitrary figures used by monax tool for now
	amount := uint64(99999999999999)
	amountBonded := uint64(9999999999)
	return GenesisSpec{
		Accounts: []TemplateAccount{{
			Name:         fmt.Sprintf("Full_%v", index),
			Amount:       &amount,
			AmountBonded: &amountBonded,
			Permissions:  []string{permission.AllString},
		},
		},
	}
}

func RootAccount(index int) GenesisSpec {
	// Inheriting from the arbitrary figures used by monax tool for now
	amount := uint64(99999999999999)
	return GenesisSpec{
		Accounts: []TemplateAccount{{
			Name:        fmt.Sprintf("Root_%v", index),
			Amount:      &amount,
			Permissions: []string{permission.AllString},
		},
		},
	}
}

func ParticipantAccount(index int) GenesisSpec {
	// Inheriting from the arbitrary figures used by monax tool for now
	amount := uint64(9999999999)
	return GenesisSpec{
		Accounts: []TemplateAccount{{
			Name:   fmt.Sprintf("Participant_%v", index),
			Amount: &amount,
			Permissions: []string{permission.SendString, permission.CallString, permission.NameString,
				permission.HasRoleString},
		}},
	}
}

func DeveloperAccount(index int) GenesisSpec {
	// Inheriting from the arbitrary figures used by monax tool for now
	amount := uint64(9999999999)
	return GenesisSpec{
		Accounts: []TemplateAccount{{
			Name:   fmt.Sprintf("Developer_%v", index),
			Amount: &amount,
			Permissions: []string{permission.SendString, permission.CallString, permission.CreateContractString,
				permission.CreateAccountString, permission.NameString, permission.HasRoleString,
				permission.RemoveRoleString},
		}},
	}
}

func ValidatorAccount(index int) GenesisSpec {
	// Inheriting from the arbitrary figures used by monax tool for now
	amount := uint64(9999999999)
	amountBonded := amount - 1
	return GenesisSpec{
		Accounts: []TemplateAccount{{
			Name:         fmt.Sprintf("Validator_%v", index),
			Amount:       &amount,
			AmountBonded: &amountBonded,
			Permissions: []string{permission.BondString},
		}},
	}
}

func MergeGenesisSpecs(genesisSpecs ...GenesisSpec) GenesisSpec {
	mergedGenesisSpec := GenesisSpec{}
	// We will deduplicate and merge global permissions flags
	permSet := make(map[string]bool)

	for _, genesisSpec := range genesisSpecs {
		// We'll overwrite chain name for later specs
		if genesisSpec.ChainName != "" {
			mergedGenesisSpec.ChainName = genesisSpec.ChainName
		}
		// Take the max genesis time
		if mergedGenesisSpec.GenesisTime == nil ||
			(genesisSpec.GenesisTime != nil && genesisSpec.GenesisTime.After(*mergedGenesisSpec.GenesisTime)) {
			mergedGenesisSpec.GenesisTime = genesisSpec.GenesisTime
		}

		for _, permString := range genesisSpec.GlobalPermissions {
			permSet[permString] = true
		}

		mergedGenesisSpec.Salt = append(mergedGenesisSpec.Salt, genesisSpec.Salt...)
		mergedGenesisSpec.Accounts = append(mergedGenesisSpec.Accounts, genesisSpec.Accounts...)
	}

	mergedGenesisSpec.GlobalPermissions = make([]string, 0, len(permSet))

	for permString := range permSet {
		mergedGenesisSpec.GlobalPermissions = append(mergedGenesisSpec.GlobalPermissions, permString)
	}

	// Make sure merged GenesisSpec is deterministic on inputs
	sort.Strings(mergedGenesisSpec.GlobalPermissions)

	return mergedGenesisSpec
}
