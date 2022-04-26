package main

import (
	"context"
	"fmt"

	"github.com/radekg/karapace-demo/opa"
)

type Address struct {
	Address1   string `avro:"address1"`
	Address2   string `avro:"address2"`
	PostalCode string `avro:"postal_code"`
	City       string `avro:"city"`
}

func (a *Address) Protected(ctx context.Context, as string) *Address {
	return &Address{
		Address1:   opa.ReadAddressFieldOrMask(ctx, "address1", as, a.Address1),
		Address2:   opa.ReadAddressFieldOrMask(ctx, "address2", as, a.Address2),
		PostalCode: opa.ReadAddressFieldOrMask(ctx, "postal_code", as, a.PostalCode),
		City:       opa.ReadAddressFieldOrMask(ctx, "city", as, a.City),
	}
}

func (a *Address) String() string {
	return fmt.Sprintf("{address1=%s, address2=%s, pcode=%s, city=%s}",
		a.Address1,
		a.Address2,
		a.PostalCode,
		a.City)
}

type Person struct {
	FirstName    string   `avro:"first_name"`
	LastName     string   `avro:"last_name"`
	EmailAddress string   `avro:"email_address"`
	HomeAddress  *Address `avro:"home_address"`
}

func (p *Person) String() string {
	return fmt.Sprintf("Person{fname=%s, lname=%s, email=%s, address=#%s}",
		p.FirstName, p.LastName, p.EmailAddress, p.HomeAddress.String())
}

func (p *Person) Protected(ctx context.Context, as string) *Person {
	return &Person{
		FirstName:    opa.ReadPersonFieldOrMask(ctx, "first_name", as, p.FirstName),
		LastName:     opa.ReadPersonFieldOrMask(ctx, "last_name", as, p.LastName),
		EmailAddress: opa.ReadPersonFieldOrMask(ctx, "email_address", as, p.EmailAddress),
		HomeAddress:  p.HomeAddress.Protected(ctx, as),
	}
}
