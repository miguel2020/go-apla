// Code generated by go generate; DO NOT EDIT.

package migration

var contractsDataSQL = `
INSERT INTO "1_contracts" (id, name, value, conditions, app_id, wallet_id, ecosystem)
VALUES
	(next_id('1_contracts'), 'AdminCondition', '// This contract is used to set "admin" rights.
// Usually the "admin" role is used for this.
// The role ID is written to the ecosystem parameter and can be changed.
// The contract requests the role ID from the ecosystem parameter and the contract checks the rights.

contract AdminCondition {
	conditions {
		if EcosysParam("founder_account") != $key_id {
			warning "Sorry, you do not have access to this action."
		}

		// var role_id int
		// role_id = EcosysParam("role_admin")
		// RoleAccess(role_id)
	}
}
', 'ContractConditions("MainCondition")', '%[5]d', %[2]d, '%[1]d'),
	(next_id('1_contracts'), 'DeveloperCondition', '// This contract is used to set "developer" rights.
// Usually the "developer" role is used for this.
// The role ID is written to the ecosystem parameter and can be changed.
// The contract requests the role ID from the ecosystem parameter and the contract checks the rights.

contract DeveloperCondition {
	conditions {
		if EcosysParam("founder_account") != $key_id {
			warning "Sorry, you do not have access to this action."
		}

		// var role_id int
		// role_id = EcosysParam("role_developer")
		// RoleAccess(role_id)
	}
}
', 'ContractConditions("MainCondition")', '%[5]d', %[2]d, '%[1]d'),
	(next_id('1_contracts'), 'MainCondition', 'contract MainCondition {
	conditions {
		if EcosysParam("founder_account")!=$key_id
		{
			warning "Sorry, you do not have access to this action."
		}
	}
}
', 'ContractConditions("MainCondition")', '%[5]d', %[2]d, '%[1]d');
`