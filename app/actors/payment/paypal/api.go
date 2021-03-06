package paypal

import (
	"fmt"

	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ottemo/commerce/api"
	"github.com/ottemo/commerce/app"
	"github.com/ottemo/commerce/env"
	"github.com/ottemo/commerce/utils"

	"github.com/ottemo/commerce/app/models/checkout"
	"github.com/ottemo/commerce/app/models/order"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	service := api.GetRestService()

	service.GET("paypal/success", APIReceipt)
	service.GET("paypal/cancel", APIDecline)

	return nil
}

// CompleteTransaction makes NVP request to PayPal for a given purchase order using PayPal token and payer id
//   - refer to https://developer.paypal.com/docs/classic/api/NVPAPIOverview/ for details
func CompleteTransaction(orderInstance order.InterfaceOrder, token string, payerID string) (map[string]string, error) {

	// getting order information
	//--------------------------
	grandTotal := orderInstance.GetGrandTotal()
	shippingPrice := orderInstance.GetShippingAmount()

	// getting request param values
	//-----------------------------
	user := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathUser))
	password := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathPass))
	signature := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathSignature))
	action := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathAction))

	amount := fmt.Sprintf("%.2f", grandTotal)
	shippingAmount := fmt.Sprintf("%.2f", shippingPrice)
	itemAmount := fmt.Sprintf("%.2f", grandTotal-shippingPrice)

	description := "Purchase%20for%20%24" + fmt.Sprintf("%.2f", grandTotal)
	custom := orderInstance.GetID()

	// making NVP request
	//-------------------
	requestParams := "USER=" + user +
		"&PWD=" + password +
		"&SIGNATURE=" + signature +
		"&METHOD=DoExpressCheckoutPayment" +
		"&VERSION=78" +
		"&PAYMENTREQUEST_0_PAYMENTACTION=" + action +
		"&PAYMENTREQUEST_0_AMT=" + amount +
		"&PAYMENTREQUEST_0_SHIPPINGAMT=" + shippingAmount +
		"&PAYMENTREQUEST_0_ITEMAMT=" + itemAmount +
		"&PAYMENTREQUEST_0_DESC=" + description +
		"&PAYMENTREQUEST_0_CUSTOM=" + custom +
		"&PAYMENTREQUEST_0_CURRENCYCODE=USD" +
		"&PAYERID=" + payerID +
		"&TOKEN=" + token

	nvpGateway := paymentPayPalExpress[
		ConstPaymentPayPalNvp][
		utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathPayPalExpressGateway))]

	request, err := http.NewRequest("GET", nvpGateway+"?"+requestParams, nil)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// reading/decoding response from NVP
	//-----------------------------------
	responseData, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	urlGETParams := make(map[string]string)
	urlParsedParams, err := url.ParseQuery(string(responseData))
	if err == nil {
		for key, value := range urlParsedParams {
			urlGETParams[key] = value[0]
		}
	}

	if urlGETParams["ACK"] != "Success" || urlGETParams["TOKEN"] == "" {
		if urlGETParams["L_ERRORCODE0"] != "" {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "2b3f49d1-b6d7-492f-ac96-854a85b64c2f", "payment confirm error "+utils.InterfaceToString(urlGETParams["L_ERRORCODE0"])+": "+"L_LONGMESSAGE0")
		}
	}

	return urlGETParams, nil
}

// APIReceipt processes PayPal receipt response
//   - "token" field should contain valid session ID
//   - refer to https://developer.paypal.com/docs/classic/api/NVPAPIOverview/ for details
func APIReceipt(context api.InterfaceApplicationContext) (interface{}, error) {
	requestData := context.GetRequestArguments()
	if !utils.KeysInMapAndNotBlank(requestData, "token", "PayerID") {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "15c2f1ca-ce41-4292-9bd3-f6232edc27c4", "token or payerID are not set")
	}
	sessionID := waitingTokens[requestData["token"]]

	waitingTokensMutex.Lock()
	delete(waitingTokens, requestData["token"])
	waitingTokensMutex.Unlock()

	sessionInstance, err := api.GetSessionByID(utils.InterfaceToString(sessionID), false)
	if sessionInstance == nil {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "cd783513-3454-4e03-bc4c-2efcb67757a0", "Wrong session ID")
	}
	err = context.SetSession(sessionInstance)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	currentCheckout, err := checkout.GetCurrentCheckout(context, true)
	if err != nil {
		return nil, err
	}

	checkoutOrder := currentCheckout.GetOrder()
	currentCart := currentCheckout.GetCart()
	if currentCart == nil {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "15f11fb8-a15e-4788-a4d3-98c41eac4caf", "Cart is not specified")
	}

	if checkoutOrder != nil {

		completeData, err := CompleteTransaction(checkoutOrder, requestData["token"], requestData["PayerID"])
		if err != nil {
			env.Log(ConstLogStorage, env.ConstLogPrefixInfo, "TRANSACTION NOT COMPLETED: "+
				"VisitorID - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", "+
				"OrderID - "+checkoutOrder.GetID()+", "+
				"TOKEN - : "+completeData["TOKEN"])

			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "af2103b1-8501-4e4b-bc1b-9086ef7c63be", "Transaction not confirmed")
		}

		result, err := currentCheckout.SubmitFinish(utils.InterfaceToMap(completeData))

		env.Log(ConstLogStorage, env.ConstLogPrefixInfo, "TRANSACTION COMPLETED: "+
			"VisitorID - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", "+
			"OrderID - "+checkoutOrder.GetID()+", "+
			"TOKEN - : "+completeData["TOKEN"])

		return api.StructRestRedirect{Result: result, Location: app.GetStorefrontURL("checkout/success/" + checkoutOrder.GetID()), DoRedirect: true}, err
	}

	return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "8d449c1c-ca34-4260-a93b-8af999c1ff04", "Checkout not exist")
}

// APIDecline processes PayPal decline response
//   - refer to https://developer.paypal.com/docs/classic/api/NVPAPIOverview/ for details
func APIDecline(context api.InterfaceApplicationContext) (interface{}, error) {
	requestData := context.GetRequestArguments()
	if !utils.KeysInMapAndNotBlank(requestData, "token", "PayerID") {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "4b0fcede-8ce7-4f4d-bee4-9cf75a427f59", "token or payerID are not set")
	}
	sessionID := waitingTokens[requestData["token"]]

	waitingTokensMutex.Lock()
	delete(waitingTokens, requestData["token"])
	waitingTokensMutex.Unlock()

	session, err := api.GetSessionByID(utils.InterfaceToString(sessionID), false)
	if session == nil {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "04eaf220-5964-4d92-821b-c60e1eb8185e", "Wrong session ID")
	}
	err = context.SetSession(session)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	currentCheckout, err := checkout.GetCurrentCheckout(context, true)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	checkoutOrder := currentCheckout.GetOrder()

	if err := checkoutOrder.SetStatus(order.ConstOrderStatusDeclined); err != nil {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "bdcbeb7f-7cf7-4ecc-b99b-779cb721ddf0", "unable to set checkout status: "+err.Error())
	}

	err = checkoutOrder.Save()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	env.Log(ConstLogStorage, env.ConstLogPrefixInfo, "Declined: "+
		"VisitorID - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", "+
		"OrderID - "+checkoutOrder.GetID()+", "+
		"TOKEN - : "+requestData["token"])

	return api.StructRestRedirect{Location: app.GetStorefrontURL("checkout"), DoRedirect: true}, nil
}
