package checkout

import (
	"time"

	"github.com/ottemo/commerce/api"
	"github.com/ottemo/commerce/env"
	"github.com/ottemo/commerce/utils"

	"github.com/ottemo/commerce/app/actors/payment/zeropay"
	"github.com/ottemo/commerce/app/models/checkout"
	"github.com/ottemo/commerce/app/models/visitor"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	service := api.GetRestService()

	service.GET("checkout", APIGetCheckout)

	// Addresses
	service.PUT("checkout/shipping/address", APISetShippingAddress)
	service.PUT("checkout/billing/address", APISetBillingAddress)

	// Shipping method
	service.GET("checkout/shipping/methods", APIGetShippingMethods)
	service.PUT("checkout/shipping/method/:method/:rate", APISetShippingMethod)

	// Payment method
	service.GET("checkout/payment/methods", APIGetPaymentMethods)
	service.PUT("checkout/payment/method/:method", APISetPaymentMethod)

	// Finalize
	service.PUT("checkout", APISetCheckoutInfo)
	service.POST("checkout/submit", APISubmitCheckout)

	return nil
}

// APIGetCheckout returns information related to current checkkout
func APIGetCheckout(context api.InterfaceApplicationContext) (interface{}, error) {

	currentCheckout, err := checkout.GetCurrentCheckout(context, false)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	result := map[string]interface{}{
		"billing_address":  nil,
		"shipping_address": nil,

		"payment_method_name": nil,
		"payment_method_code": nil,

		"shipping_method_name": nil,
		"shipping_method_code": nil,

		"shipping_rate":   nil,
		"shipping_amount": nil,

		"discounts":       nil,
		"discount_amount": nil,

		"taxes":      nil,
		"tax_amount": nil,

		"subtotal":   nil,
		"grandtotal": nil,
		"info":       nil,
	}

	if billingAddress := currentCheckout.GetBillingAddress(); billingAddress != nil {
		result["billing_address"] = billingAddress.ToHashMap()
	}

	if shippingAddress := currentCheckout.GetShippingAddress(); shippingAddress != nil {
		shippingAddressMap := shippingAddress.ToHashMap()

		if notes := utils.InterfaceToString(currentCheckout.GetInfo("notes")); notes != "" {
			shippingAddressMap["notes"] = notes
		}

		result["shipping_address"] = shippingAddressMap
	}

	if paymentMethod := currentCheckout.GetPaymentMethod(); paymentMethod != nil {
		result["payment_method_name"] = paymentMethod.GetName()
		result["payment_method_code"] = paymentMethod.GetCode()
	}

	if shippingMethod := currentCheckout.GetShippingMethod(); shippingMethod != nil {
		result["shipping_method_name"] = shippingMethod.GetName()
		result["shipping_method_code"] = shippingMethod.GetCode()
	}

	if shippingRate := currentCheckout.GetShippingRate(); shippingRate != nil {
		result["shipping_rate"] = shippingRate
	}

	result["grandtotal"] = currentCheckout.GetGrandTotal()
	result["subtotal"] = currentCheckout.GetSubtotal()

	result["shipping_amount"] = currentCheckout.GetShippingAmount()

	result["tax_amount"] = currentCheckout.GetTaxAmount()
	result["taxes"] = currentCheckout.GetTaxes()

	result["discount_amount"] = currentCheckout.GetDiscountAmount()
	result["discounts"] = currentCheckout.GetDiscounts()

	// The info map is only returned for logged out users
	infoMap := make(map[string]interface{})

	for key, value := range utils.InterfaceToMap(currentCheckout.GetInfo("*")) {
		// prevent from showing cc values in info
		if key != "cc" {
			infoMap[key] = value
		}
	}

	result["info"] = infoMap

	return result, nil
}

// APIGetPaymentMethods returns currently available payment methods
func APIGetPaymentMethods(context api.InterfaceApplicationContext) (interface{}, error) {

	currentCheckout, err := checkout.GetCurrentCheckout(context, false)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	type ResultValue struct {
		Name      string
		Code      string
		Type      string
		Tokenable bool
	}
	var result []ResultValue

	// for checkout that contain subscription items we will show only payment methods that allows to save token
	isSubscription := currentCheckout.IsSubscription()

	for _, paymentMethod := range checkout.GetRegisteredPaymentMethods() {
		if paymentMethod.IsAllowed(currentCheckout) && (!isSubscription || paymentMethod.IsTokenable(currentCheckout)) {
			result = append(result, ResultValue{Name: paymentMethod.GetName(), Code: paymentMethod.GetCode(), Type: paymentMethod.GetType(), Tokenable: paymentMethod.IsTokenable(currentCheckout)})
		}
	}

	return result, nil
}

// APIGetShippingMethods returns currently available shipping methods
func APIGetShippingMethods(context api.InterfaceApplicationContext) (interface{}, error) {

	currentCheckout, err := checkout.GetCurrentCheckout(context, false)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	type ResultValue struct {
		Name  string
		Code  string
		Rates []checkout.StructShippingRate
	}
	var result []ResultValue

	for _, shippingMethod := range checkout.GetRegisteredShippingMethods() {
		if shippingMethod.IsAllowed(currentCheckout) {
			result = append(result, ResultValue{Name: shippingMethod.GetName(), Code: shippingMethod.GetCode(), Rates: shippingMethod.GetRates(currentCheckout)})
		}
	}

	return result, nil
}

// APISetCheckoutInfo allows to specify and assign to checkout extra information
func APISetCheckoutInfo(context api.InterfaceApplicationContext) (interface{}, error) {

	currentCheckout, err := checkout.GetCurrentCheckout(context, true)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	for key, value := range requestData {
		err := currentCheckout.SetInfo(key, value)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	// updating session
	if err := checkout.SetCurrentCheckout(context, currentCheckout); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "174cee3e-d82f-42c6-841a-7b26eedde0be", err.Error())
	}

	return "ok", nil
}

// checkoutObtainAddress is an internal usage function used create and validate address
//   - address data supposed to be in request content
func checkoutObtainAddress(data interface{}) (visitor.InterfaceVisitorAddress, error) {

	var err error
	var currentVisitorID string
	var addressData map[string]interface{}

	switch context := data.(type) {
	case api.InterfaceApplicationContext:
		currentVisitorID = utils.InterfaceToString(context.GetSession().Get(visitor.ConstSessionKeyVisitorID))
		addressData, err = api.GetRequestContentAsMap(context)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	case map[string]interface{}:
		addressData = context
	default:
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "995c0d71-5465-4003-9ed2-347bcf6255c5", "unknown address data type")
	}

	// checking for address id was specified, if it was - making sure it correct
	if addressID, present := addressData["id"]; present {

		// loading specified address by id
		visitorAddress, err := visitor.LoadVisitorAddressByID(utils.InterfaceToString(addressID))
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		// checking address owner is current visitor

		if currentVisitorID != "" && visitorAddress.GetVisitorID() != currentVisitorID {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "bef27714-4ac5-4705-b59a-47c8e0bc5aa4", "address id is not related to current visitor")
		}

		return visitorAddress, nil
	}

	visitorAddressModel, err := checkout.ValidateAddress(addressData)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// setting address owner to current visitor (for sure)
	if currentVisitorID != "" {
		if err := visitorAddressModel.Set("visitor_id", currentVisitorID); err != nil {
			_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "103248e6-00ee-463d-b877-24e7cd56bb98", err.Error())
		}
	}

	// if address id was specified it means that address was changed, so saving it
	// new address we are not saving as if could be temporary address
	if (visitorAddressModel.GetID() != "" || currentVisitorID != "") && utils.InterfaceToBool(addressData["save"]) {
		err = visitorAddressModel.Save()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	return visitorAddressModel, nil
}

// APISetShippingAddress specifies shipping address for a current checkout
func APISetShippingAddress(context api.InterfaceApplicationContext) (interface{}, error) {
	currentCheckout, err := checkout.GetCurrentCheckout(context, true)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	address, err := checkoutObtainAddress(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = currentCheckout.SetShippingAddress(address)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	requestContents, _ := api.GetRequestContentAsMap(context)

	if notes, present := requestContents["notes"]; present {
		if err := currentCheckout.SetInfo("notes", notes); err != nil {
			_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "be9cc966-ad69-4115-b1d3-269ae46f071b", err.Error())
		}
	}

	// updating session
	if err := checkout.SetCurrentCheckout(context, currentCheckout); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "03f68537-5165-4b19-a42e-cf6b03ca81f3", err.Error())
	}

	return address.ToHashMap(), nil
}

// APISetBillingAddress specifies billing address for a current checkout
func APISetBillingAddress(context api.InterfaceApplicationContext) (interface{}, error) {
	currentCheckout, err := checkout.GetCurrentCheckout(context, true)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	address, err := checkoutObtainAddress(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = currentCheckout.SetBillingAddress(address)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// updating session
	if err := checkout.SetCurrentCheckout(context, currentCheckout); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "674ca503-4035-403d-96ea-3a5a952916cf", err.Error())
	}

	return address.ToHashMap(), nil
}

// APISetPaymentMethod assigns payment method to current checkout
//   - "method" argument specifies requested payment method (it should be available for a meaning time)
func APISetPaymentMethod(context api.InterfaceApplicationContext) (interface{}, error) {

	currentCheckout, err := checkout.GetCurrentCheckout(context, true)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// looking for payment method
	for _, paymentMethod := range checkout.GetRegisteredPaymentMethods() {
		if paymentMethod.GetCode() == context.GetRequestArgument("method") {
			if paymentMethod.IsAllowed(currentCheckout) {

				// updating checkout payment method
				err := currentCheckout.SetPaymentMethod(paymentMethod)
				if err != nil {
					return nil, env.ErrorDispatch(err)
				}

				// checking for additional info
				contentValues, _ := api.GetRequestContentAsMap(context)
				for key, value := range contentValues {
					if err := currentCheckout.SetInfo(key, value); err != nil {
						_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "7e5e7fc9-a6f5-41c7-84e0-b029535da7dd", err.Error())
					}
				}

				// visitor event for setting payment method
				eventData := map[string]interface{}{"session": context.GetSession(), "paymentMethod": paymentMethod, "checkout": currentCheckout}
				env.Event("api.checkout.setPayment", eventData)

				// updating session
				if err := checkout.SetCurrentCheckout(context, currentCheckout); err != nil {
					_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "15b1fad6-f514-45c5-b7ae-1c97060331dc", err.Error())
				}

				return "ok", nil
			}
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "becea4f9-39b6-4710-b96a-e7ff262823dc", "payment method not allowed")
		}
	}

	return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "4d39abcd-3724-44fe-818e-c8a99da4d780", "payment method not found")
}

// APISetShippingMethod assigns shipping method and shipping rate to current checkout
//   - "method" argument specifies requested shipping method (it should be available for a meaning time)
//   - "rate" argument specifies requested shipping rate (it should be available and belongs to shipping method)
func APISetShippingMethod(context api.InterfaceApplicationContext) (interface{}, error) {

	currentCheckout, err := checkout.GetCurrentCheckout(context, true)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// looking for shipping method
	for _, shippingMethod := range checkout.GetRegisteredShippingMethods() {
		if shippingMethod.GetCode() == context.GetRequestArgument("method") {
			if shippingMethod.IsAllowed(currentCheckout) {

				// looking for shipping rate
				for _, shippingRate := range shippingMethod.GetRates(currentCheckout) {
					if shippingRate.Code == context.GetRequestArgument("rate") {

						err := currentCheckout.SetShippingMethod(shippingMethod)
						if err != nil {
							return nil, env.ErrorDispatch(err)
						}

						err = currentCheckout.SetShippingRate(shippingRate)
						if err != nil {
							return nil, env.ErrorDispatch(err)
						}

						// updating session
						if err := checkout.SetCurrentCheckout(context, currentCheckout); err != nil {
							_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "59982278-bef4-4abc-ba1b-4cb9a3e300e2", err.Error())
						}

						return "ok", nil
					}
				}

			} else {
				return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "d7fb6ff2-b914-467b-bf56-b8d2bea472ef", "shipping method not allowed")
			}
		}
	}

	return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "c589e0f9-4e4a-4691-8f43-a045aaff48c2", "shipping method and/or rate were not found")
}

// checkoutObtainToken is an internal usage function used to create or load credit card for visitor
func checkoutObtainToken(currentCheckout checkout.InterfaceCheckout, creditCardInfo map[string]interface{}) (visitor.InterfaceVisitorCard, error) {

	// make sure we have a visitor
	currentVisitor := currentCheckout.GetVisitor()
	currentVisitorID := ""
	if currentVisitor != nil {
		currentVisitorID = currentVisitor.GetID()
	}
	if currentVisitorID == "" {
		err := env.ErrorNew(ConstErrorModule, 10, "c9e46525-77f6-4add-b286-efeb8a63f4d1", "user not logged in, don't attempt to save a token")
		return nil, err
	}

	// if we were passed a token rowID, make sure it belongs to the user
	if creditCardID := utils.GetFirstMapValue(creditCardInfo, "id", "_id"); creditCardID != nil {

		// loading specified credit card by id
		visitorCard, err := visitor.LoadVisitorCardByID(utils.InterfaceToString(creditCardID))
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		// checking address owner is current visitor
		if visitorCard.GetVisitorID() != currentVisitorID {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "3b5446ef-dd70-4bdc-8d30-817cb7b48d05", "credit card id is not related to current visitor")
		}

		return visitorCard, nil
	}

	// we weren't passed a token, start generating one
	paymentMethod := currentCheckout.GetPaymentMethod()
	if !paymentMethod.IsTokenable(currentCheckout) {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "5b05cc24-2184-47cc-b2dc-77cb41035698", "for selected payment method credit card can't be saved")
	}

	billingName := ""
	if add := currentCheckout.GetBillingAddress(); add != nil {
		billingName = add.GetFirstName() + " " + add.GetLastName()
	}

	// put required key to create token from payment method using only zero amount authorize
	paymentInfo := map[string]interface{}{
		checkout.ConstPaymentActionTypeKey: checkout.ConstPaymentActionTypeCreateToken,
		"cc": creditCardInfo,
		"extra": map[string]interface{}{
			"email":        currentVisitor.GetEmail(),
			"visitor_id":   currentVisitorID,
			"billing_name": billingName,
		},
	}
	// contains creditCardLastFour, creditCardType, responseMessage, responseResult, transactionID, creditCardExp
	paymentResult, err := paymentMethod.Authorize(nil, paymentInfo)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	authorizeCardResult := utils.InterfaceToMap(paymentResult)
	if !utils.KeysInMapAndNotBlank(authorizeCardResult, "transactionID", "creditCardLastFour") {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "22e17290-56f3-452a-8d54-18d5a9eb2833", "transaction can't be obtained")
	}

	// create visitor card and fill required fields
	//---------------------------------
	visitorCardModel, err := visitor.GetVisitorCardModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// override credit card info with provided from payment info
	// TODO: payment should have interface method that return predefined struct for 0 authorize
	creditCardInfo["token_id"] = authorizeCardResult["transactionID"]
	creditCardInfo["payment"] = paymentMethod.GetCode()
	creditCardInfo["customer_id"] = authorizeCardResult["customerID"]
	creditCardInfo["type"] = authorizeCardResult["creditCardType"]
	creditCardInfo["number"] = authorizeCardResult["creditCardLastFour"]
	creditCardInfo["expiration_date"] = authorizeCardResult["creditCardExp"] // mmyy
	creditCardInfo["token_updated"] = time.Now()
	creditCardInfo["created_at"] = time.Now()

	// filling new instance with request provided data
	for attribute, value := range creditCardInfo {
		err := visitorCardModel.Set(attribute, value)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	// setting credit card owner to current visitor (for sure)
	if err := visitorCardModel.Set("visitor_id", currentVisitorID); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "46466d05-75a6-4da3-9e5d-f366868b33ba", err.Error())
	}

	// save card info if checkbox is checked on frontend
	if utils.InterfaceToBool(creditCardInfo["save"]) {
		err = visitorCardModel.Save()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	return visitorCardModel, nil
}

// APISubmitCheckout submits current checkout and creates a new order base on it
func APISubmitCheckout(context api.InterfaceApplicationContext) (interface{}, error) {

	// preparations
	//--------------
	currentCheckout, err := checkout.GetCurrentCheckout(context, true)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// Handle custom information set in case of one request submit
	if customInfo := utils.GetFirstMapValue(requestData, "custom_info"); customInfo != nil {
		for key, value := range utils.InterfaceToMap(customInfo) {
			if err := currentCheckout.SetInfo(key, value); err != nil {
				_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "190134b1-cc8b-4744-a896-532d39d7fbbf", err.Error())
			}
		}
	}

	if err := currentCheckout.SetInfo("session_id", context.GetSession().GetID()); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "6405e87f-8d5c-4fa6-828d-da0e5b77f15a", err.Error())
	}
	currentVisitorID := utils.InterfaceToString(context.GetSession().Get(visitor.ConstSessionKeyVisitorID))

	addressInfoToAddress := func(addressInfo interface{}) (visitor.InterfaceVisitorAddress, error) {
		var addressData map[string]interface{}

		switch typedValue := addressInfo.(type) {
		case map[string]interface{}:
			typedValue["visitor_id"] = currentVisitorID
			addressData = typedValue
		case string:
			addressData = map[string]interface{}{"visitor_id": currentVisitorID, "id": typedValue}
		}
		return checkoutObtainAddress(addressData)
	}

	// checking for specified shipping address
	//-----------------------------------------
	if shippingAddressInfo := utils.GetFirstMapValue(requestData, "shipping_address", "shippingAddress", "ShippingAddress"); shippingAddressInfo != nil {
		address, err := addressInfoToAddress(shippingAddressInfo)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
		if err := currentCheckout.SetShippingAddress(address); err != nil {
			_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "63009038-bde5-4147-aad2-7586e9d0736c", err.Error())
		}
	}

	// checking for specified billing address
	//----------------------------------------
	if billingAddressInfo := utils.GetFirstMapValue(requestData, "billing_address", "billingAddress", "BillingAddress"); billingAddressInfo != nil {
		address, err := addressInfoToAddress(billingAddressInfo)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
		if err := currentCheckout.SetBillingAddress(address); err != nil {
			_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "73a6599a-23f2-4066-ac29-b02ed627f51f", err.Error())
		}
	}

	// checking for specified payment method
	//---------------------------------------
	if specifiedPaymentMethod := utils.GetFirstMapValue(requestData, "payment_method", "paymentMethod"); specifiedPaymentMethod != nil {
		var found bool
		for _, paymentMethod := range checkout.GetRegisteredPaymentMethods() {
			if paymentMethod.GetCode() == specifiedPaymentMethod {
				if paymentMethod.IsAllowed(currentCheckout) {
					err := currentCheckout.SetPaymentMethod(paymentMethod)
					if err != nil {
						return nil, env.ErrorDispatch(err)
					}
					found = true
					break
				}
				return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "bd07849e-8789-4316-924c-9c754efbc348", "payment method not allowed")
			}
		}

		if !found {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "b8384a47-8806-4a54-90fc-cccb5e958b4e", "payment method not found")
		}
	}

	currentPaymentMethod := currentCheckout.GetPaymentMethod()
	// set ZeroPayment method for checkout without payment method
	if currentPaymentMethod == nil || (currentCheckout.GetGrandTotal() == 0 && currentPaymentMethod.GetCode() != zeropay.ConstPaymentZeroPaymentCode) {
		for _, paymentMethod := range checkout.GetRegisteredPaymentMethods() {
			if zeropay.ConstPaymentZeroPaymentCode == paymentMethod.GetCode() {
				if paymentMethod.IsAllowed(currentCheckout) {
					err := currentCheckout.SetPaymentMethod(paymentMethod)
					if err != nil {
						return nil, env.ErrorDispatch(err)
					}
				}
			}
		}
	}

	// checking for specified shipping method
	//----------------------------------------
	specifiedShippingMethod := utils.GetFirstMapValue(requestData, "shipping_method", "shipppingMethod")
	specifiedShippingMethodRate := utils.GetFirstMapValue(requestData, "shipppingRate", "shipping_rate")

	if specifiedShippingMethod != nil && specifiedShippingMethodRate != nil {
		var methodFound, rateFound bool

		for _, shippingMethod := range checkout.GetRegisteredShippingMethods() {
			if shippingMethod.GetCode() == specifiedShippingMethod {
				if shippingMethod.IsAllowed(currentCheckout) {
					methodFound = true

					for _, shippingRate := range shippingMethod.GetRates(currentCheckout) {
						if shippingRate.Code == specifiedShippingMethodRate {
							err = currentCheckout.SetShippingMethod(shippingMethod)
							if err != nil {
								return nil, env.ErrorDispatch(err)
							}
							if err := currentCheckout.SetShippingRate(shippingRate); err != nil {
								_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "9ac7f59a-3e65-46a2-bbb8-344de50dd8e6", err.Error())
							}
							if err != nil {
								return nil, env.ErrorDispatch(err)
							}

							rateFound = true
							break
						}
					}
					break
				}
			}
		}

		if !methodFound || !rateFound {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "279a645c-6a03-44de-95c0-2651a51440fa", "shipping method and/or rate were not found")
		}
	}

	// Now that checkout is about to submit we want to see if we can turn our cc info into a token
	// cc info can be used directly from post body
	specifiedCreditCard := utils.GetFirstMapValue(requestData, "cc", "ccInfo", "creditCardInfo")
	if specifiedCreditCard == nil {
		specifiedCreditCard = currentCheckout.GetInfo("cc")
		// if credit card was already handled and saved to cc info, we will pass this handling
		if creditCard, ok := specifiedCreditCard.(visitor.InterfaceVisitorCard); ok && creditCard != nil {
			specifiedCreditCard = nil
		}
	}

	// Add handle for credit card post action in one request, it would bind credit card object to a cc key in checkout info
	if specifiedCreditCard != nil {
		// credit card wouldn't be saved to checkout if it's not response to current visitor/payment
		creditCard, err := checkoutObtainToken(currentCheckout, utils.InterfaceToMap(specifiedCreditCard))
		if err != nil {
			// in  this case raw cc will be set to checkout info and used by payment method
			if err := currentCheckout.SetInfo("cc", specifiedCreditCard); err != nil {
				_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "c0ba05ec-1300-455a-8858-cf91e4bb615d", err.Error())
			}
			_ = env.ErrorDispatch(err)
		} else {
			if err := currentCheckout.SetInfo("cc", creditCard); err != nil {
				_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "9b36aae8-9e43-45d1-84ec-0dd1f13d3e18", err.Error())
			}
		}
	}

	return currentCheckout.Submit()
}
