package bybit

// func TestWsClient_SubscribeMarket(t *testing.T) {
// 	cli, err := NewWsClient(context.Background(), func(bs []byte) error {
// 		t.Logf("received -> %+v", string(bs))
// 		return nil
// 	})
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if err = cli.SubscribeMarket("CELRUSDT"); err != nil {
// 		t.Error(err)
// 	}
// 	time.Sleep(time.Minute)
// }

// func TestWsClient_SubscribeSpotMarket(t *testing.T) {
// 	cli, err := NewWsClient(context.Background(), func(bs []byte) error {
// 		t.Logf("received -> %+v", string(bs))
// 		return nil
// 	}, SetSpotV2())
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if err = cli.SubscribeSpotMarket("BTCUSDT", "ETHUSDT"); err != nil {
// 		t.Error(err)
// 	}
// 	time.Sleep(time.Minute)
// }

// func TestWsClient_SubscribeSpotTrade(t *testing.T) {
// 	cli, err := NewWsClient(context.Background(), func(bs []byte) error {
// 		t.Logf("received -> %+v", string(bs))
// 		return nil
// 	}, SetSpotV2())
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if err = cli.SubscribeSpotTrade("BTCUSDT"); err != nil {
// 		t.Error(err)
// 	}
// 	time.Sleep(time.Minute)
// }
