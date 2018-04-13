package viabtc

import (
	"fmt"

	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"time"

	"github.com/go-errors/errors"
)

// Config is an structure which holds configurable parameters of
// ViaBTC client.
type Config struct {
	// Host points out to the main client server.
	Host string

	// Port denotes the port on which client server is listening for
	// incoming requests.
	Port int
}

// Client is the programmatic connector to the core exchange client,
// currently is written to interact using http requests to access rpc
// function point, but in future could be rewritten to use unix sockets or
// even use embedded C code.
type Client struct {
	httpClient *http.Client
	url        string
}

// NewClient creates new instance of ViaBTC client client.
func NewClient(cfg *Config) *Client {
	httpUrl := fmt.Sprintf("http://%v:%v", cfg.Host, cfg.Port)

	return &Client{
		httpClient: &http.Client{},
		url:        httpUrl,
	}
}

// makeRPCCall is a helper which is used to execute client remote
// procedure call using http post request, with encoded parameters in a request
// body. On return the rpc response object is populated with data which is
// specific for ever call.
func (e *Client) makeRPCCall(method string, params interface{},
	rpcResp interface{}) error {

	args, err := extractArguments(params)
	if err != nil {
		return errors.Errorf("unable to extract arguments: %v", err)
	}

	rpcReq := &request{
		Method: method,
		Params: args,
		ID:     int32(time.Now().Unix()),
	}

	data, err := json.Marshal(rpcReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", e.url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, rpcResp); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("status code: %v", resp.StatusCode)
	}

	return nil
}

// Accounts returns available and frozen balances of user for every
// supported by client currency.
func (e *Client) BalanceQuery(params *BalanceQueryRequest) (
	BalanceQueryResponse, error) {

	type Response struct {
		baseResponse
		Result BalanceQueryResponse
	}

	response := &Response{}
	err := e.makeRPCCall("balance.query", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// BalanceUpdate updates balance of the user, it is used by the
// client itself to update balances on order matching, and it is used by
// backend subsystem to deposit and withdraw funds.
//
// NOTE: If request is sent with the same action id it will be discarded by
// the client.
func (e *Client) BalanceUpdate(params *BalanceUpdateRequest) (
	*BalanceUpdateResponse, error) {

	type Response struct {
		baseResponse
		Result *BalanceUpdateResponse
	}

	response := &Response{}
	err := e.makeRPCCall("balance.update", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// BalanceHistory returns the history of all funds changes we have been done
// with the user chosen asset.
func (e *Client) BalanceHistory(params *BalanceHistoryRequest) (
	*BalanceHistoryResponse, error) {

	type Response struct {
		baseResponse
		Result *BalanceHistoryResponse
	}

	response := &Response{}
	err := e.makeRPCCall("balance.history", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// AssetList returns the list of assets and its calculation precious, i.e.
// how much decimal places is preserved in client when operating with asset.
func (e *Client) AssetList(params *AssetListRequest) (
	*AssetListResponse, error) {

	type Response struct {
		baseResponse
		Result *AssetListResponse
	}

	response := &Response{}
	err := e.makeRPCCall("asset.list", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// AssetSummary returns the aggregated information for all accounts about
// overall available volume corresponding to assets, overall freezed volume,
// and how much accounts have the asset available.
func (e *Client) AssetSummary(params *AssetSummaryRequest) (
	*AssetSummaryResponse, error) {

	type Response struct {
		baseResponse
		Result *AssetSummaryResponse
	}

	response := &Response{}
	err := e.makeRPCCall("asset.summary", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderPutLimit puts the order on the market with fixed price and amount, if
// there is enough volume for specified price the order will be waiting for
// opposite order to come.
func (e *Client) OrderPutLimit(params *OrderPutLimitRequest) (
	*OrderPutLimitResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderPutLimitResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.put_limit", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderPutMarket puts the order on the market. As far as price is not fixed
// and the goal of the order is to be fully executed than if market has enough
// volume for executing the order it will be fully handled. But in this case
// the average price might be much lower than the market price.
func (e *Client) OrderPutMarket(params *OrderPutMarketRequest) (
	*OrderPutMarketResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderPutMarketResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.put_market", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderCancel cancels the order of specific user on the market.
func (e *Client) OrderCancel(params *OrderCancelRequest) (
	*OrderCancelResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderCancelResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.cancel", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderBook by the given market and side returns all available on
// current moment orders.
func (e *Client) OrderBook(params *OrderBookRequest) (
	*OrderBookResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderBookResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.book", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderDepth returns the overall volume for each available price, also if
// interval is specified than volume within the interval will be combined.
func (e *Client) OrderDepth(params *OrderDepthRequest) (
	*OrderDepthResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderDepthResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.depth", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderPending returns the user's pending orders with their detailed
// information.
func (e *Client) OrderPending(params *OrderPendingRequest) (
	*OrderPendingResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderPendingResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.pending", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderPendingDetail returns the detailed information about specific order.
func (e *Client) OrderPendingDetail(params *OrderPendingDetailRequest) (
	*OrderPendingDetailResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderPendingDetailResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.pending_detail", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderDeals returns the information about the operations which has been
// applied to order in order to fulfil it. When the order is completely
// fulfilled the number of deals might be [1, inf).
func (e *Client) OrderDeals(params *OrderDealsRequest) (
	*OrderDealsResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderDealsResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.deals", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderFinished returns the information about user's finished orders.
func (e *Client) OrderFinished(params *OrderFinishedRequest) (
	*OrderFinishedResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderFinishedResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.finished", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderFinishedDetail returns the detailed information about specific
// finished order.
func (e *Client) OrderFinishedDetail(params *OrderFinishedDetailRequest) (
	*OrderFinishedDetailResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderFinishedDetailResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.finished_detail", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketLast returns last market price.
func (e *Client) MarketLast(params *MarketLastRequest) (
	*string, error) {

	type Response struct {
		baseResponse
		Result *string
	}

	response := &Response{}
	err := e.makeRPCCall("market.last", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketSummary returns the aggregated information for all accounts about
// overall available orders corresponding to market.
func (e *Client) MarketSummary(params *MarketSummaryRequest) (
	*MarketSummaryResponse, error) {

	type Response struct {
		baseResponse
		Result *MarketSummaryResponse
	}

	response := &Response{}
	err := e.makeRPCCall("market.summary", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketList return the information about the market's calculation precious,
// and minimum amount of order.
func (e *Client) MarketList(params *MarketListRequest) (
	*MarketListResponse, error) {

	type Response struct {
		baseResponse
		Result *MarketListResponse
	}

	response := &Response{}
	err := e.makeRPCCall("market.list", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketDeals returns the information about
func (e *Client) MarketDeals(params *MarketDealsRequest) (
	MarketDealsResponse, error) {

	type Response struct {
		baseResponse
		Result MarketDealsResponse
	}

	response := &Response{}
	err := e.makeRPCCall("market.deals", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketUserDeals returns the information about deals which were made by
// user. Deal is the result of orders matching.
func (e *Client) MarketUserDeals(params *MarketUserDealsRequest) (
	*MarketUserDealsResponse, error) {

	type Response struct {
		baseResponse
		Result *MarketUserDealsResponse
	}

	response := &Response{}
	err := e.makeRPCCall("market.user_deals", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketKLine returns the information about the market withing preset
// interval of time. The number of requests is determined as (e - s) / i, where
// e - end time, s - start time, i - interval.
func (e *Client) MarketKLine(params *MarketKLineRequest) (
	MarketKLineResponse, error) {

	type Response struct {
		baseResponse
		Result MarketKLineResponse
	}

	response := &Response{}
	err := e.makeRPCCall("market.kline", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketStatus returns the status of the market within given period of time.
func (e *Client) MarketStatus(params *MarketStatusRequest) (
	*MarketStatusResponse, error) {

	type Response struct {
		baseResponse
		Result *MarketStatusResponse
	}

	response := &Response{}
	err := e.makeRPCCall("market.status", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketStatusToday returns the information about the market within the
// current day.
func (e *Client) MarketStatusToday(params *MarketStatusTodayRequest) (
	*MarketStatusTodayResponse, error) {

	type Response struct {
		baseResponse
		Result *MarketStatusTodayResponse
	}

	response := &Response{}
	err := e.makeRPCCall("market.status_today", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}
