package checkout

import (
	"github.com/ottemo/foundation/env"
)

// Package variables
//------------------

var (
	ShippingMethods = make([]I_ShippingMethod, 0)
	PaymentMethods  = make([]I_PaymentMethod, 0)

	Taxes     = make([]I_Tax, 0)
	Discounts = make([]I_Discount, 0)
)

// Managing routines
//------------------

// register new shipping method to system
func RegisterShippingMethod(shippingMethod I_ShippingMethod) error {
	for _, registeredMethod := range ShippingMethods {
		if registeredMethod == shippingMethod {
			return env.ErrorNew("shipping method already registered")
		}
	}

	ShippingMethods = append(ShippingMethods, shippingMethod)

	return nil
}

// register new payment method to system
func RegisterPaymentMethod(paymentMethod I_PaymentMethod) error {
	for _, registeredMethod := range PaymentMethods {
		if registeredMethod == paymentMethod {
			return env.ErrorNew("payment method already registered")
		}
	}

	PaymentMethods = append(PaymentMethods, paymentMethod)

	return nil
}

// register new tax calculator in system
func RegisterTax(tax I_Tax) error {
	for _, registeredTax := range Taxes {
		if registeredTax == tax {
			return env.ErrorNew("tax already registered")
		}
	}

	Taxes = append(Taxes, tax)

	return nil
}

// register new discount calculator in system
func RegisterDiscount(discount I_Discount) error {
	for _, registeredDiscount := range Discounts {
		if registeredDiscount == discount {
			return env.ErrorNew("discount already registered")
		}
	}

	Discounts = append(Discounts, discount)

	return nil
}

// returns list of registered shipping methods
func GetRegisteredShippingMethods() []I_ShippingMethod {
	return ShippingMethods
}

// returns list of registered payment methods
func GetRegisteredPaymentMethods() []I_PaymentMethod {
	return PaymentMethods
}

// returns list of registered tax calculators
func GetRegisteredTaxes() []I_Tax {
	return Taxes
}

// returns list of registered tax calculators
func GetRegisteredDiscounts() []I_Discount {
	return Discounts
}
