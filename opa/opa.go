package opa

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/rego"
)

const opaModule = `
package karapace.demo

default can_read_protected_address = false

restricted_address_fields := {"address1", "address2", "postal_code"}

can_read_protected_address {
	not is_hr
    not restricted_address_fields[input.field]
}

can_read_protected_address {
	is_hr
    restricted_address_fields[_]
}

default can_read_protected_person = false
restricted_person_fields := {"email_address"}

can_read_protected_person {
	not is_hr
    not restricted_person_fields[input.field]
}

can_read_protected_person {
	is_hr
    restricted_person_fields[_]
}

is_hr {
	input.subject.groups[_] = "hr"
}
`

func ReadAddressFieldOrMask(ctx context.Context, field, role, defaultVal string) string {
	if canReadAddressField(ctx, field, role) {
		return defaultVal
	}
	return "**********"
}

func ReadPersonFieldOrMask(ctx context.Context, field, role, defaultVal string) string {
	if canReadPersonField(ctx, field, role) {
		return defaultVal
	}
	return "**********"
}

func canReadAddressField(ctx context.Context, field, role string) bool {
	return canReadProtectedField(ctx, "can_read_protected_address", field, role)
}

func canReadPersonField(ctx context.Context, field, role string) bool {
	return canReadProtectedField(ctx, "can_read_protected_person", field, role)
}

func canReadProtectedField(ctx context.Context, expression, field, role string) bool {
	// suboptimal ofc, we should compile this rule once...
	query, err := rego.New(
		// if it's executed with a binding:
		// rego.Query("x = data.karapace.demo.can_read_protected_address"),
		// it will not eval to Allowed(), Allowed() returns the expected value
		// only if there are no bindings!
		rego.Query(fmt.Sprintf("data.karapace.demo.%s", expression)),
		rego.Module("karapace-demo.rego", opaModule),
	).PrepareForEval(ctx)

	if err != nil {
		return false
	}

	results, err := query.Eval(ctx, rego.EvalInput(map[string]interface{}{
		"field": field,
		"subject": map[string]interface{}{
			"groups": []interface{}{"employee", role},
		},
	}))

	if err != nil {
		return false
	}

	return results.Allowed()
}
