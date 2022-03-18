package shopee

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type CheckoutValidationError struct{ json jsoniter.Any }

func (c CheckoutValidationError) Code() int   { return c.json.Get("error").ToInt() }
func (c CheckoutValidationError) Msg() string { return c.json.Get("error_msg").ToString() }

func (c CheckoutValidationError) Error() string {
	return fmt.Sprintf("checkout validation error: code=%d, validation_code=%d, %s", c.Code(), c.ValidationCode(), c.Msg())
}

func (c CheckoutValidationError) ValidationCode() int {
	return c.json.Get("data", "validation_error").ToInt()
}

func (c Client) ValidateCheckout(item CheckoutableItem) error {
	type obj = map[string]interface{}
	type arr = []interface{}
	resp, err := c.Client.R().
		SetBody(obj{
			"shop_orders": arr{
				obj{
					"shop_info": obj{"shop_id": item.ShopID()},
					"item_infos": arr{
						obj{
							"item_id":  item.ItemID(),
							"model_id": item.ChosenModel().ModelID(),
							"quantity": 1,
						},
					},
					"buyer_address": nil,
				},
			},
		}).
		Post("/api/v4/pdp/buy_now/validate_checkout")
	if err != nil {
		return err
	}

	json := jsoniter.Get(resp.Body())
	if json.Get("error").ToInt() != 0 || json.Get("data", "validation").ToInt() != 0 {
		return CheckoutValidationError{json}
	}

	return nil
}

type PlaceOrderError struct{ json jsoniter.Any }

func (p PlaceOrderError) Type() string { return p.json.Get("error").ToString() }
func (p PlaceOrderError) Msg() string  { return p.json.Get("error_msg").ToString() }

func (p PlaceOrderError) Error() string {
	return fmt.Sprintf("%s: %s", p.Type(), p.Msg())
}

// all exported fields are required
type CheckoutParams struct {
	Addr        AddressInfo
	Item        CheckoutableItem
	PaymentData PaymentChannelData
	Logistic    LogisticChannelInfo
	ts          int64
}

func (c CheckoutParams) WithTimestamp(ts int64) CheckoutParams {
	c.ts = ts
	return c
}

//go:embed porder.json
var porderPayload []byte

func (c Client) PlaceOrder(params CheckoutParams) error {
	if params.ts == 0 {
		return errors.New("no timestamp in params")
	}

	var data map[string]interface{}
	if err := jsoniter.Unmarshal(porderPayload, &data); err != nil {
		return err
	}

	item := params.Item
	logistic := params.Logistic
	type p = []interface{}
	for _, field := range [...]struct {
		p
		v interface{}
	}{
		{p{"timestamp"}, params.ts},
		{p{"selected_payment_channel_data"}, params.PaymentData},
		{p{"shoporders", 0, "shop", "shopid"}, item.ShopID()},
		{p{"shoporders", 0, "items", 0, "itemid"}, item.ItemID()},
		{p{"shoporders", 0, "items", 0, "modelid"}, item.ChosenModel().ModelID()},
		{p{"shoporders", 0, "items", 0, "shopid"}, item.ShopID()},
		{p{"shoporders", 0, "items", 0, "price"}, item.ChosenModel().Price()},
		{p{"shoporders", 0, "items", 0, "name"}, item.Name()},
		{p{"shoporders", 0, "items", 0, "model_name"}, item.ChosenModel().Name()},
		{p{"shoporders", 0, "items", 0, "categories", 0, "catids"}, item.CatIDs()},
		{p{"shipping_orders", 0, "selected_logistic_channelid"}, logistic.ChannelID()},
		{p{"shipping_orders", 0, "buyer_address_data", "addressid"}, params.Addr.ID()},
		{p{"checkout_price_data", "merchandise_subtotal"}, item.ChosenModel().Price()},
		{p{"checkout_price_data", "shipping_subtotal_before_discount"}, logistic.PriceMax()},
		{p{"checkout_price_data", "shipping_subtotal"}, logistic.PriceMax()},
		{p{"checkout_price_data", "total_payable"}, item.ChosenModel().Price() + logistic.PriceMax() + 250000000},
		{p{"shoporders", 0, "order_total_without_shipping"}, item.ChosenModel().Price()},
		{p{"shoporders", 0, "order_total"}, item.ChosenModel().Price() + logistic.PriceMax()},
		{p{"shipping_orders", 0, "order_total"}, item.ChosenModel().Price() + logistic.PriceMax()},
		{p{"shipping_orders", 0, "order_total_without_shipping"}, item.ChosenModel().Price()},
		{p{"shipping_orders", 0, "shipping_fee"}, logistic.PriceMax()},
	} {
		setJsonField(data, field.p, field.v)
	}

	resp, err := c.Client.R().
		SetBody(data).
		Post("/api/v4/checkout/place_order")
	if err != nil {
		return err
	}

	json := jsoniter.Get(resp.Body())
	if json.Get("error").ToString() != "" {
		return PlaceOrderError{json}
	}
	return nil
}

//go:embed cgetq.json
var cgetqPayload []byte

func (c Client) CheckoutGetQuick(params CheckoutParams) (CheckoutParams, error) {
	var data map[string]interface{}
	if err := jsoniter.Unmarshal(cgetqPayload, &data); err != nil {
		return CheckoutParams{}, err
	}

	ts := time.Now().Unix()

	item := params.Item
	type p = []interface{}
	for _, field := range []struct {
		p
		v interface{}
	}{
		{p{"timestamp"}, ts},
		{p{"shoporders", 0, "shop", "shopid"}, item.ShopID()},
		{p{"shoporders", 0, "items", 0, "itemid"}, item.ItemID()},
		{p{"shoporders", 0, "items", 0, "modelid"}, item.ChosenModel().ModelID()},
	} {
		setJsonField(data, field.p, field.v)
	}

	_, err := c.Client.R().
		SetBody(data).
		Post("/api/v4/checkout/get_quick")
	if err != nil {
		return CheckoutParams{}, err
	}

	return params.WithTimestamp(ts), nil
}

//go:embed cget.json
var cgetPayload []byte

func (c Client) CheckoutGet(params CheckoutParams) (CheckoutParams, error) {
	var data map[string]interface{}
	if err := jsoniter.Unmarshal(cgetPayload, &data); err != nil {
		return CheckoutParams{}, err
	}

	ts := time.Now().Unix()

	item := params.Item
	type p = []interface{}
	for _, field := range [...]struct {
		p
		v interface{}
	}{
		{p{"timestamp"}, ts},
		{p{"shoporders", 0, "shop", "shopid"}, item.ShopID()},
		{p{"shoporders", 0, "items", 0, "itemid"}, item.ItemID()},
		{p{"shoporders", 0, "items", 0, "modelid"}, item.ChosenModel().ModelID()},
		{p{"selected_payment_channel_data"}, params.PaymentData},
		{p{"shipping_orders", 0, "buyer_address_data", "addressid"}, params.Addr.ID()},
		{p{"shipping_orders", 0, "selected_logistic_channelid"}, params.Logistic.ChannelID()},
	} {
		setJsonField(data, field.p, field.v)
	}

	_, err := c.Client.R().
		SetBody(data).
		Post("/api/v4/checkout/get")
	if err != nil {
		return CheckoutParams{}, err
	}

	return params.WithTimestamp(ts), nil
}

func setJsonField(json interface{}, path []interface{}, v interface{}) {
	for i, accessor := range path[:len(path)-1] {
		switch jsontyp := json.(type) {
		case map[string]interface{}:
			json = jsontyp[accessor.(string)]
		case []interface{}:
			json = jsontyp[accessor.(int)]
		default:
			log.Printf("path: %#+v\n", path)
			log.Printf("json: %#+v\n", json)
			panic(fmt.Sprint("invalid accessor: ", accessor, " at index ", i))
		}
	}
	field := path[len(path)-1]
	switch json := json.(type) {
	case map[string]interface{}:
		json[field.(string)] = v
	case []interface{}:
		json[field.(int)] = v
	}
}
