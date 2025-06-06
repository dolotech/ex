package bybit

import (
	"errors"
	"fmt"
	"net/http"
)

func (b *Client) CreateOrderV2(side string, orderType string, price float64,
	qty int, timeInForce string, takeProfit float64, stopLoss float64, reduceOnly bool,
	closeOnTrigger bool, orderLinkID string, symbol string) (result OrderV2, err error) {
	var cResult CreateOrderV2Result
	params := map[string]interface{}{}
	params["side"] = side
	params["symbol"] = symbol
	params["order_type"] = orderType
	params["qty"] = qty
	if price > 0 {
		params["price"] = price
	}
	params["time_in_force"] = timeInForce
	if takeProfit > 0 {
		params["take_profit"] = takeProfit
	}
	if stopLoss > 0 {
		params["stop_loss"] = stopLoss
	}
	if reduceOnly {
		params["reduce_only"] = true
	}
	if closeOnTrigger {
		params["close_on_trigger"] = true
	}
	if orderLinkID != "" {
		params["order_link_id"] = orderLinkID
	}
	var resp []byte
	resp, err = b.SignedRequest(http.MethodPost, "v2/private/order/create", params, &cResult)
	if err != nil {
		return
	}
	if cResult.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", cResult.RetMsg, string(resp))
		return
	}
	result = cResult.Result
	return
}

// CreateLinearOrder 創建活動委托單
// POST /private/linear/order/create
// side				*	string	方向
// symbol			*	string	合約類型
// order_type		*	string	委托單價格類型
// qty				*	number	委托數量(BTC)
// price				number	委托價格。如果是下限價單，該參數為必填. 在沒有倉位時，做多的委托價格需高於市價的10%、低於1百萬。如有倉位時則需優於強平價。價格增減最小單位請參考交易對接口響應中的price_filter字段
// time_in_force	*	string	執行策略
// reduce_only		*	bool	什麽是 reduce-only order?,true-平倉 false-開倉,ture時止盈止損設置不生效
// close_on_trigger	*	bool	什麽是 close on trigger order?只會減少您的倉位而不會增加您的倉位。如果當平倉委托被觸發時，賬戶上的余額不足，那麽該合約的其他委托將被取消或者降低委托數量。使用此選項可以確保您的止損單被用於減倉而非加倉。
// order_link_id		string	機構自定義訂單ID, 最大長度36位，且同一機構下自定義ID不可重復
// take_profit			number	止盈價格，僅開倉時生效
// stop_loss			number	止損價格，僅開倉時生效
// tp_trigger_by		string	止盈激活價格類型，默認為LastPrice
// sl_trigger_by		string	止損激活價格類型，默認為LastPrice
// position_idx			integer	Position idx, 用於在不同倉位模式下標識倉位。 如果是單向持倉模式， 該參數為必填:
// 									0-單向持倉
// 									1-雙向持倉Buy
// 									2-雙向持倉Sell
func (b *Client) CreateLinearOrder(param *CreateOrderParam) (result Order, err error) {
	var cResult CreateOrderResult
	params := map[string]interface{}{}
	params["side"] = param.Side
	params["symbol"] = param.Symbol
	params["order_type"] = param.OrderType
	params["qty"] = param.Qty
	params["time_in_force"] = param.TimeInForce
	params["reduce_only"] = param.ReduceOnly
	params["close_on_trigger"] = param.CloseOnTrigger
	params["position_idx"] = param.PositionIdx
	if param.Price > 0 {
		params["price"] = param.Price
	}
	if param.OrderLinkID != "" {
		params["order_link_id"] = param.OrderLinkID
	}
	if param.TakeProfit != 0 {
		params["take_profit"] = param.TakeProfit
	}
	if param.StopLoss != 0 {
		params["stop_loss"] = param.StopLoss
	}
	if param.TpTriggerBy != "" {
		params["tp_trigger_by"] = param.TpTriggerBy
	}
	if param.SlTriggerBy != "" {
		params["sl_trigger_by"] = param.SlTriggerBy
	}
	var resp []byte
	resp, err = b.SignedRequest(http.MethodPost, "private/linear/order/create", params, &cResult)
	if err != nil {
		return
	}
	if cResult.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", cResult.RetMsg, string(resp))
		return
	}
	result = cResult.Result
	return
}

// GetLinearOrders 查詢活動委托
// GET /private/linear/order/list
// order_id				string	訂單ID
// order_link_id		string	機構自定義訂單ID, 最大長度36位，且同一機構下自定義ID不可重復
// symbol			*	string	合約類型
// order				string	按創建時間排序（默認升序）
// page					integer	頁碼.默認取第一頁,最大50
// limit				integer	每頁數量, 最大50. 默認每頁20條
// order_status			string	指定訂單狀態查詢訂單列表。不傳該參數則默認查詢所有狀態訂單。該參數支持多狀態查詢，狀態之間用英文逗號分割。
func (b *Client) GetLinearOrders(orderID, orderLinkID string, page int, limit int, order string, orderStatus string, symbol string) (result []Order, err error) {
	var cResult OrderListResult
	params := map[string]interface{}{}
	params["symbol"] = symbol
	if orderID != "" {
		params["order_id"] = orderID
	}
	if orderLinkID != "" {
		params["order_link_id"] = orderLinkID
	}
	if order != "" {
		params["order"] = order
	}
	params["page"] = page
	params["limit"] = limit
	if orderStatus != "" {
		params["order_status"] = orderStatus
	}
	var resp []byte
	resp, err = b.SignedRequest(http.MethodGet, "private/linear/order/list", params, &cResult)
	if err != nil {
		return
	}
	if cResult.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", cResult.RetMsg, string(resp))
		return
	}

	result = cResult.Result.Data
	return
}

// CreateOrder 创建委托单
// symbol: 产品类型, 有效选项:BTCUSD,ETHUSD (BTCUSD ETHUSD)
// side: 方向, 有效选项:Buy, Sell (Buy Sell)
// orderType: Limit/Market
// price: 委托价格, 在没有仓位时，做多的委托价格需高于市价的10%、低于1百万。如有仓位时则需优于强平价。单笔价格增减最小单位为0.5。
// qty: 委托数量, 单笔最大1百万
// timeInForce: 执行策略, 有效选项:GoodTillCancel,ImmediateOrCancel,FillOrKill,PostOnly
// reduceOnly: 只减仓
// symbol: 产品类型, 有效选项:BTCUSD,ETHUSD (BTCUSD ETHUSD)
func (b *Client) CreateOrder(side string, orderType string, price float64, qty int, timeInForce string, reduceOnly bool, symbol string) (result Order, err error) {
	var cResult CreateOrderResult
	params := map[string]interface{}{}
	params["side"] = side
	params["symbol"] = symbol
	params["order_type"] = orderType
	params["qty"] = qty
	params["price"] = price
	params["time_in_force"] = timeInForce
	if reduceOnly {
		params["reduce_only"] = true
	}
	var resp []byte
	resp, err = b.SignedRequest(http.MethodPost, "open-api/order/create", params, &cResult)
	if err != nil {
		return
	}
	if cResult.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", cResult.RetMsg, string(resp))
		return
	}
	result = cResult.Result
	return
}

func (b *Client) ReplaceOrder(symbol string, orderID string, qty int, price float64) (result Order, err error) {
	var cResult ReplaceOrderResult
	params := map[string]interface{}{}
	params["order_id"] = orderID
	params["symbol"] = symbol
	if qty > 0 {
		params["p_r_qty"] = qty
	}
	if price > 0 {
		params["p_r_price"] = price
	}
	var resp []byte
	resp, err = b.SignedRequest(http.MethodPost, "open-api/order/replace", params, &cResult)
	if err != nil {
		return
	}
	if cResult.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", cResult.RetMsg, string(resp))
		return
	}
	result.OrderID = cResult.Result.OrderID
	return
}

// CreateStopOrder 创建条件委托单
// https://github.com/bybit-exchange/bybit-official-api-docs/blob/master/zh_cn/rest_api.md#open-apistop-ordercreatepost
// symbol: 产品类型, 有效选项:BTCUSD,ETHUSD (BTCUSD ETHUSD)
// side: 方向, 有效选项:Buy, Sell (Buy Sell)
// orderType: Limit/Market
// price: 委托价格, 在没有仓位时，做多的委托价格需高于市价的10%、低于1百万。如有仓位时则需优于强平价。单笔价格增减最小单位为0.5。
// qty: 委托数量, 单笔最大1百万
// basePrice: 当前市价。用于和stop_px值进行比较，确定当前条件委托是看空到stop_px时触发还是看多到stop_px触发。主要是用来标识当前条件单预期的方向
// stopPx: 条件委托下单时市价
// triggerBy: 触发价格类型. 默认为上一笔成交价格
// timeInForce: 执行策略, 有效选项:GoodTillCancel,ImmediateOrCancel,FillOrKill,PostOnly
// reduceOnly: 只减仓
// symbol: 产品类型, 有效选项:BTCUSD,ETHUSD (BTCUSD ETHUSD)
func (b *Client) CreateStopOrder(side string, orderType string, price float64, basePrice float64, stopPx float64,
	qty int, triggerBy string, timeInForce string, reduceOnly bool, symbol string) (result Order, err error) {
	var cResult CreateOrderResult
	params := map[string]interface{}{}
	params["side"] = side
	params["symbol"] = symbol
	params["order_type"] = orderType
	params["qty"] = qty
	if price > 0 {
		params["price"] = price
	}
	params["base_price"] = basePrice
	params["stop_px"] = stopPx
	params["time_in_force"] = timeInForce
	if reduceOnly {
		params["reduce_only"] = true
	}
	if triggerBy != "" {
		params["trigger_by"] = triggerBy
	}
	var resp []byte
	resp, err = b.SignedRequest(http.MethodPost, "open-api/stop-order/create", params, &cResult)
	if err != nil {
		return
	}
	if cResult.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", cResult.RetMsg, string(resp))
		return
	}
	result = cResult.Result
	return
}

// GetOrders 查询活动委托
// symbol
// orderID: 订单ID
// orderLinkID: 机构自定义订单ID
// sort: 排序字段，默认按创建时间排序 (created_at cum_exec_qty qty last_exec_price price cum_exec_value cum_exec_fee)
// order: 升序降序， 默认降序 (desc asc)
// page: 页码，默认取第一页数据
// limit: 一页数量，一页默认展示20条数据
func (b *Client) GetOrders(sort string, order string, page int,
	limit int, orderStatus string, symbol string) (result []Order, err error) {
	return b.getOrders("", "", sort, order, page, limit, orderStatus, symbol)
}

// getOrders 查询活动委托
// symbol
// orderID: 订单ID
// orderLinkID: 机构自定义订单ID
// sort: 排序字段，默认按创建时间排序 (created_at cum_exec_qty qty last_exec_price price cum_exec_value cum_exec_fee)
// order: 升序降序， 默认降序 (desc asc)
// page: 页码，默认取第一页数据
// limit: 一页数量，一页默认展示20条数据
func (b *Client) getOrders(orderID string, orderLinkID string, sort string, order string, page int,
	limit int, orderStatus string, symbol string) (result []Order, err error) {
	var cResult OrderListResult

	if limit == 0 {
		limit = 20
	}

	params := map[string]interface{}{}
	params["symbol"] = symbol
	if orderID != "" {
		params["order_id"] = orderID
	}
	if orderLinkID != "" {
		params["order_link_id"] = orderLinkID
	}
	if sort != "" {
		params["sort"] = sort
	}
	if order != "" {
		params["order"] = order
	}
	params["page"] = page
	params["limit"] = limit
	if orderStatus != "" {
		params["order_status"] = orderStatus
	}
	var resp []byte
	resp, err = b.SignedRequest(http.MethodGet, "open-api/order/list", params, &cResult)
	if err != nil {
		return
	}
	if cResult.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", cResult.RetMsg, string(resp))
		return
	}

	result = cResult.Result.Data
	return
}

// GetStopOrders 查询条件委托单
// orderID: 条件委托单ID
// orderLinkID: 机构自定义订单ID
// order: 排序字段为created_at,升序降序，默认降序 (desc asc )
// page: 页码，默认取第一页数据
// stopOrderStatus 条件单状态: Untriggered: 等待市价触发条件单; Triggered: 市价已触发条件单; Cancelled: 取消; Active: 条件单触发成功且下单成功; Rejected: 条件触发成功但下单失败
// limit: 一页数量，默认一页展示20条数据;最大支持50条每页
func (b *Client) GetStopOrders(orderID string, orderLinkID string, stopOrderStatus string, order string,
	page int, limit int, symbol string) (result GetStopOrdersResult, err error) {

	if limit == 0 {
		limit = 20
	}

	params := map[string]interface{}{}
	params["symbol"] = symbol
	if orderID != "" {
		params["stop_order_id"] = orderID
	}
	if orderLinkID != "" {
		params["order_link_id"] = orderLinkID
	}
	if stopOrderStatus != "" {
		params["stop_order_status"] = stopOrderStatus
	}
	if order != "" {
		params["order"] = order
	}
	params["page"] = page
	params["limit"] = limit
	var resp []byte
	resp, err = b.SignedRequest(http.MethodGet, "open-api/stop-order/list", params, &result)
	if err != nil {
		return
	}
	if result.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", result.RetMsg, string(resp))
		return
	}

	return
}

// GetOrderByID
func (b *Client) GetOrderByID(orderID string, orderLinkID string, symbol string) (result OrderV2, err error) {
	var cResult QueryOrderResult

	params := map[string]interface{}{}
	params["symbol"] = symbol
	if orderID != "" {
		params["order_id"] = orderID
	}
	if orderLinkID != "" {
		params["order_link_id"] = orderLinkID
	}
	var resp []byte
	resp, err = b.SignedRequest(http.MethodGet, "v2/private/order", params, &cResult)
	if err != nil {
		return
	}
	if cResult.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", cResult.RetMsg, string(resp))
		return
	}

	result = cResult.Result
	return
}

// GetOrderByOrderLinkID ...
func (b *Client) GetOrderByOrderLinkID(orderLinkID string, symbol string) (result Order, err error) {
	var orders []Order
	orders, err = b.getOrders("", orderLinkID, "", "", 0, 20, "", symbol)
	if err != nil {
		return
	}
	if len(orders) != 1 {
		err = errors.New("not found")
		return
	}
	result = orders[0]
	return
}

// CancelOrder 撤销活动委托单
// orderID: 活动委托单ID, 数据来自创建活动委托单返回的订单唯一ID
// symbol:
func (b *Client) CancelOrder(orderID string, symbol string) (result Order, err error) {
	var cResult CancelOrderResult
	params := map[string]interface{}{}
	params["symbol"] = symbol
	params["order_id"] = orderID
	var resp []byte
	resp, err = b.SignedRequest(http.MethodPost, "open-api/order/cancel", params, &cResult)
	if err != nil {
		return
	}
	if cResult.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", cResult.RetMsg, string(resp))
		return
	}

	result = cResult.Result
	return
}

// CancelOrder 撤销活动委托单
// orderID: 活动委托单ID, 数据来自创建活动委托单返回的订单唯一ID
// symbol:
func (b *Client) CancelOrderV2(orderID string, orderLinkID string, symbol string) (result OrderV2, err error) {
	var cResult CancelOrderV2Result
	params := map[string]interface{}{}
	params["symbol"] = symbol
	if orderID != "" {
		params["order_id"] = orderID
	}
	if orderLinkID != "" {
		params["order_link_id"] = orderLinkID
	}
	var resp []byte
	resp, err = b.SignedRequest(http.MethodPost, "v2/private/order/cancel", params, &cResult)
	if err != nil {
		return
	}
	if cResult.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", cResult.RetMsg, string(resp))
		return
	}

	result = cResult.Result
	return
}

// CancelAllOrder Cancel All Active Orders
func (b *Client) CancelAllOrder(symbol string) (result []OrderV2, err error) {
	var cResult CancelAllOrderV2Result
	params := map[string]interface{}{}
	params["symbol"] = symbol
	var resp []byte
	resp, err = b.SignedRequest(http.MethodPost, "v2/private/order/cancelAll", params, &cResult)
	if err != nil {
		return
	}
	if cResult.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", cResult.RetMsg, string(resp))
		return
	}

	result = cResult.Result
	return
}

// CancelStopOrder 撤销活动条件委托单
// orderID: 活动条件委托单ID, 数据来自创建活动委托单返回的订单唯一ID
// symbol:
func (b *Client) CancelStopOrder(orderID string, symbol string) (result Order, err error) {
	var cResult CancelOrderResult
	params := map[string]interface{}{}
	params["symbol"] = symbol
	params["stop_order_id"] = orderID
	var resp []byte
	resp, err = b.SignedRequest(http.MethodPost, "open-api/stop-order/cancel", params, &cResult)
	if err != nil {
		return
	}
	if cResult.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", cResult.RetMsg, string(resp))
		return
	}

	result = cResult.Result
	return
}

// CancelAllStopOrders 撤消全部条件委托单
// symbol:
func (b *Client) CancelAllStopOrders(symbol string) (result []StopOrderV2, err error) {
	var cResult CancelStopOrdersV2Result
	params := map[string]interface{}{}
	params["symbol"] = symbol
	var resp []byte
	resp, err = b.SignedRequest(http.MethodPost, "v2/private/stop-order/cancelAll", params, &cResult)
	if err != nil {
		return
	}
	if cResult.RetCode != 0 {
		err = fmt.Errorf("%v body: [%v]", cResult.RetMsg, string(resp))
		return
	}

	result = cResult.Result
	return
}
