package order

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ottemo/commerce/db"
	"github.com/ottemo/commerce/env"
	"github.com/ottemo/commerce/utils"

	"github.com/ottemo/commerce/app/models/cart"
	"github.com/ottemo/commerce/app/models/checkout"
	"github.com/ottemo/commerce/app/models/order"
	"github.com/ottemo/commerce/app/models/product"
	"github.com/ottemo/commerce/app/models/visitor"
)

// GetItems returns order items for current order
func (it *DefaultOrder) GetItems() []order.InterfaceOrderItem {
	var result []order.InterfaceOrderItem

	var keys []int
	for key := range it.Items {
		keys = append(keys, key)
	}

	sort.Ints(keys)

	for _, key := range keys {
		result = append(result, it.Items[key])
	}

	return result

}

// AddItem adds line item to current order, or returns error
func (it *DefaultOrder) AddItem(productID string, qty int, productOptions map[string]interface{}) (order.InterfaceOrderItem, error) {

	productModel, err := product.LoadProductByID(productID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = productModel.ApplyOptions(productOptions)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	newOrderItem := new(DefaultOrderItem)
	newOrderItem.OrderID = it.GetID()

	err = newOrderItem.Set("product_id", productModel.GetID())
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = newOrderItem.Set("qty", qty)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = newOrderItem.Set("options", productModel.GetOptions())
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = newOrderItem.Set("name", productModel.GetName())
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = newOrderItem.Set("sku", productModel.GetSku())
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = newOrderItem.Set("short_description", productModel.GetShortDescription())
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = newOrderItem.Set("price", productModel.GetPrice())
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = newOrderItem.Set("weight", productModel.GetWeight())
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	it.maxIdx++
	newOrderItem.idx = it.maxIdx
	it.Items[newOrderItem.idx] = newOrderItem

	return newOrderItem, nil
}

// RemoveAllItems removes all items from current order, or returns error
func (it *DefaultOrder) RemoveAllItems() error {
	for itemIdx := range it.Items {
		err := it.RemoveItem(itemIdx)
		if err != nil {
			return env.ErrorDispatch(err)
		}
	}
	return nil
}

// RemoveItem removes line item from current order, or returns error
func (it *DefaultOrder) RemoveItem(itemIdx int) error {
	if orderItem, present := it.Items[itemIdx]; present {

		dbEngine := db.GetDBEngine()
		if dbEngine == nil {
			return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "54410b67-aff0-418f-ad76-6453a2d6fed6", "can't get DB engine")
		}

		orderItemsCollection, err := dbEngine.GetCollection(ConstCollectionNameOrderItems)
		if err != nil {
			return env.ErrorDispatch(err)
		}

		err = orderItemsCollection.DeleteByID(orderItem.GetID())
		if err != nil {
			return env.ErrorDispatch(err)
		}

		delete(it.Items, itemIdx)

		return nil
	}

	return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "1bd2f0f9-a457-43d1-a9db-e05b1aa7e1d2", "can't find index "+utils.InterfaceToString(itemIdx))
}

// NewIncrementID assigns new unique increment id to order
func (it *DefaultOrder) NewIncrementID() error {
	lastIncrementIDMutex.Lock()

	lastIncrementID++
	it.IncrementID = fmt.Sprintf(ConstIncrementIDFormat, lastIncrementID)

	if err := env.GetConfig().SetValue(ConstConfigPathLastIncrementID, lastIncrementID); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "43b9a95b-1cd7-4ede-a685-978eda1c136f", err.Error())
	}

	lastIncrementIDMutex.Unlock()

	return nil
}

// GetIncrementID returns increment id of order
func (it *DefaultOrder) GetIncrementID() string {
	return it.IncrementID
}

// SetIncrementID sets increment id to order
func (it *DefaultOrder) SetIncrementID(incrementID string) error {
	it.IncrementID = incrementID

	return nil
}

// CalculateTotals recalculates order Subtotal and GrandTotal
func (it *DefaultOrder) CalculateTotals() error {

	it.GrandTotal = utils.RoundPrice(it.GetSubtotal() + it.GetShippingAmount() + it.GetTaxAmount() + it.GetDiscountAmount())

	return nil
}

// GetSubtotal returns subtotal of order
func (it *DefaultOrder) GetSubtotal() float64 {
	var subtotal float64
	for _, orderItem := range it.Items {
		subtotal += utils.RoundPrice(orderItem.GetPrice() * float64(orderItem.GetQty()))
	}
	it.Subtotal = utils.RoundPrice(subtotal)

	return it.Subtotal
}

// GetGrandTotal returns grand total of order
func (it *DefaultOrder) GetGrandTotal() float64 {
	return it.GrandTotal
}

// GetDiscountAmount returns discount amount applied to order
func (it *DefaultOrder) GetDiscountAmount() float64 {
	return it.Discount
}

// GetDiscounts returns discount applied to order
func (it *DefaultOrder) GetDiscounts() []order.StructDiscount {
	return it.Discounts
}

// GetTaxAmount returns tax amount applied to order
func (it *DefaultOrder) GetTaxAmount() float64 {
	return it.TaxAmount
}

// GetTaxes returns taxes applied to order
func (it *DefaultOrder) GetTaxes() []order.StructTaxRate {
	return it.Taxes
}

// GetShippingAmount returns order shipping cost
func (it *DefaultOrder) GetShippingAmount() float64 {
	return it.ShippingAmount
}

// GetShippingMethod returns shipping method for order
func (it *DefaultOrder) GetShippingMethod() string {
	return it.ShippingMethod
}

// GetPaymentMethod returns payment method used for order
func (it *DefaultOrder) GetPaymentMethod() string {
	return it.PaymentMethod
}

// GetShippingAddress returns shipping address for order
func (it *DefaultOrder) GetShippingAddress() visitor.InterfaceVisitorAddress {
	addressModel, _ := visitor.GetVisitorAddressModel()
	if err := addressModel.FromHashMap(it.ShippingAddress); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "ffcb0fb8-51b6-44df-b0f9-6377a765656c", err.Error())
	}

	return addressModel
}

// GetBillingAddress returns billing address for order
func (it *DefaultOrder) GetBillingAddress() visitor.InterfaceVisitorAddress {
	addressModel, _ := visitor.GetVisitorAddressModel()
	if err := addressModel.FromHashMap(it.BillingAddress); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "e21f8402-2642-4abd-9c80-b4b5d4af80e2", err.Error())
	}

	return addressModel
}

// GetStatus returns current order status
func (it *DefaultOrder) GetStatus() string {
	return it.Status
}

// SetStatus changes status for current order
//   - if status change no supposing stock operations, order instance will not be saved automatically
func (it *DefaultOrder) SetStatus(newStatus string) error {
	var err error

	// cases with no actions
	if it.Status == newStatus || newStatus == "" {
		return nil
	}

	// changing status
	oldStatus := it.Status
	it.Status = newStatus

	// if order new status is "new" or "declined" - returning items to stock, otherwise taking them from
	if newStatus == order.ConstOrderStatusDeclined || newStatus == order.ConstOrderStatusNew || newStatus == order.ConstOrderStatusCancelled {

		if oldStatus != order.ConstOrderStatusNew && oldStatus != order.ConstOrderStatusDeclined && oldStatus != order.ConstOrderStatusCancelled && oldStatus != "" {
			err = it.Rollback()
		}

	} else {

		// taking items from stock
		if oldStatus == order.ConstOrderStatusDeclined || oldStatus == order.ConstOrderStatusCancelled || oldStatus == order.ConstOrderStatusNew || oldStatus == "" {
			err = it.Proceed()
		}
	}

	return env.ErrorDispatch(err)
}

// Proceed subtracts order items from stock, changes status to new if status was not set yet, saves order
func (it *DefaultOrder) Proceed() error {

	if it.Status == "" {
		it.Status = order.ConstOrderStatusNew
	}

	var err error
	stockManager := product.GetRegisteredStock()
	if stockManager != nil {
		for _, orderItem := range it.GetItems() {
			options := orderItem.GetOptions()

			currProductOptions := make(map[string]interface{})
			for optionName, optionValue := range options {
				if optionValue, ok := optionValue.(map[string]interface{}); ok {
					if value, present := optionValue["value"]; present {
						currProductOptions[optionName] = value
					}
				}
			}

			err := stockManager.UpdateProductQty(orderItem.GetProductID(), currProductOptions, -1*orderItem.GetQty())
			if err != nil {
				return env.ErrorDispatch(err)
			}

		}
	}

	// checking order's incrementID, if not set - assigning new one
	if it.GetIncrementID() == "" {
		err = it.NewIncrementID()
		if err != nil {
			return env.ErrorDispatch(err)
		}
	}

	err = it.Save()
	if err != nil {
		return env.ErrorDispatch(err)
	}

	eventData := map[string]interface{}{"order": it}
	env.Event("order.proceed", eventData)

	return nil
}

// Rollback returns order items to stock, modifieds the order status to declined
// if status was not set yet, then saves order
func (it *DefaultOrder) Rollback() error {
	if it.Status == "" {
		it.Status = order.ConstOrderStatusDeclined
	}

	var err error
	stockManager := product.GetRegisteredStock()
	if stockManager != nil {
		for _, orderItem := range it.GetItems() {
			options := orderItem.GetOptions()

			currProductOptions := make(map[string]interface{})
			for optionName, optionValue := range options {
				if optionValue, ok := optionValue.(map[string]interface{}); ok {
					if value, present := optionValue["value"]; present {
						currProductOptions[optionName] = value
					}
				}
			}
			err := stockManager.UpdateProductQty(orderItem.GetProductID(), currProductOptions, orderItem.GetQty())
			if err != nil {
				return env.ErrorDispatch(err)
			}
		}
	}

	err = it.Save()
	if err != nil {
		return env.ErrorDispatch(err)
	}

	eventData := map[string]interface{}{"order": it}
	env.Event("order.rollback", eventData)

	return nil
}

// DuplicateOrder used to create checkout from order with changing params
// main params for duplication: sessionID, paymentMethod, shippingMethod
func (it *DefaultOrder) DuplicateOrder(params map[string]interface{}) (interface{}, error) {

	duplicateCheckout, err := checkout.GetCheckoutModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// set visitor basic info
	visitorID := it.Get("visitor_id")
	if visitorID != "" {
		if err := duplicateCheckout.Set("VisitorID", visitorID); err != nil {
			_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "d00a8e8c-3ba5-4736-81b0-f06b17e88dbe", err.Error())
		}
	}

	if err := duplicateCheckout.SetInfo("customer_email", it.Get("customer_email")); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "63fda490-a695-462c-b51f-55bdeaecd1e5", err.Error())
	}
	if err := duplicateCheckout.SetInfo("customer_name", it.Get("customer_name")); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "2b829e06-f8e1-4ac1-ba9e-20314e044810", err.Error())
	}

	// set billing and shipping address
	shippingAddress := it.GetShippingAddress().ToHashMap()
	if err := duplicateCheckout.Set("ShippingAddress", shippingAddress); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "96e398da-6be2-453e-9d61-a1dad52918ac", err.Error())
	}

	billingAddress := it.GetBillingAddress().ToHashMap()
	if err := duplicateCheckout.Set("BillingAddress", billingAddress); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "ffd351cd-57b6-4a60-9650-65e62e7784ad", err.Error())
	}

	// convert order Item object to cart
	currentCart, err := cart.GetCartModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	for _, orderItem := range it.GetItems() {
		itemOptions := make(map[string]interface{})

		for option, value := range orderItem.GetOptions() {
			optionMap := utils.InterfaceToMap(value)
			if optionValue, present := optionMap["value"]; present {
				itemOptions[option] = optionValue
			}
		}

		_, err = currentCart.AddItem(orderItem.GetProductID(), orderItem.GetQty(), itemOptions)
		if err != nil {
			_ = env.ErrorDispatch(err)
		}
	}

	err = currentCart.ValidateCart()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = currentCart.Save()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = duplicateCheckout.SetCart(currentCart)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// check shipping method for availability
	var methodFind, rateFind bool

	orderShipping := strings.Split(it.GetShippingMethod(), "/")
	for _, shippingMethod := range checkout.GetRegisteredShippingMethods() {
		if orderShipping[0] == shippingMethod.GetCode() {
			if shippingMethod.IsAllowed(duplicateCheckout) {
				methodFind = true

				for _, shippingRates := range shippingMethod.GetRates(duplicateCheckout) {
					if orderShipping[1] == shippingRates.Code {
						err := duplicateCheckout.SetShippingRate(shippingRates)
						if err != nil {
							_ = env.ErrorDispatch(err)
							continue
						}

						err = duplicateCheckout.SetShippingMethod(shippingMethod)
						if err != nil {
							_ = env.ErrorDispatch(err)
							methodFind = false
							continue
						}

						rateFind = true
						break
					}
				}
			}
		}
		if methodFind && rateFind {
			break
		}
	}

	// check payment method for availability
	orderPayment := it.GetPaymentMethod()
	for _, paymentMethod := range checkout.GetRegisteredPaymentMethods() {
		if orderPayment == paymentMethod.GetCode() {
			if paymentMethod.IsAllowed(duplicateCheckout) {
				err := duplicateCheckout.SetPaymentMethod(paymentMethod)
				if err != nil {
					_ = env.ErrorDispatch(err)
					continue
				}

				break
			}
		}
	}

	err = duplicateCheckout.SetInfo("cc", it.Get("payment_info"))
	if err != nil {
		_ = env.ErrorDispatch(err)
	}

	return duplicateCheckout, nil
}
