package shopee

type PaymentChannelData map[string]interface{}

type PaymentChannelOption struct{ Name, OptionInfo string }

type PaymentChannel struct {
	Name      string
	Options   []PaymentChannelOption
	ApplyFunc func(opt PaymentChannelOption) PaymentChannelData
}

// wrapper to ApplyFunc, for channel that doesn't have option.
// returns nil if channel has option
func (p PaymentChannel) Apply() PaymentChannelData {
	if len(p.Options) > 0 {
		return nil
	}
	return p.ApplyFunc(PaymentChannelOption{})
}

// wrapper to ApplyFunc
func (p PaymentChannel) ApplyOpt(opt PaymentChannelOption) PaymentChannelData {
	return p.ApplyFunc(opt)
}

var (
	ShopeePay = PaymentChannel{
		Name: "ShopeePay",
		ApplyFunc: func(PaymentChannelOption) PaymentChannelData {
			return PaymentChannelData{
				"channel_id": 8001400,
				"version":    2,
			}
		},
	}

	COD = PaymentChannel{
		Name: "COD (Bayar di Tempat)",
		ApplyFunc: func(PaymentChannelOption) PaymentChannelData {
			return PaymentChannelData{
				"channel_id": 89000,
				"version":    1,
			}
		},
	}

	TransferBank = PaymentChannel{
		Name: "Transfer Bank",
		Options: []PaymentChannelOption{
			{"SeaBank (Dicek Otomatis)", "89052007"},
			{"Bank BCA (Dicek Otomatis)", "89052001"},
			{"Bank Mandiri (Dicek Otomatis)", "89052002"},
			{"Bank BNI (Dicek Otomatis)", "89052003"},
			{"Bank BRI (Dicek Otomatis)", "89052004"},
			{"Bank Syariah Indonesia (BSI) (Dicek Otomatis)", "89052005"},
			{"Bank Permata (Dicek Otomatis)", "89052006"},
			{"Bank lainnya (Dicek Otomatis)", "89052902"},
		},
		ApplyFunc: func(opt PaymentChannelOption) PaymentChannelData {
			return PaymentChannelData{
				"channel_id":               8005200,
				"channel_item_option_info": PaymentChannelData{"option_info": opt.OptionInfo},
				"version":                  2,
			}
		},
	}

	Alfamart = PaymentChannel{
		Name: "Alfamart / Alfamidi / Dan+Dan",
		ApplyFunc: func(PaymentChannelOption) PaymentChannelData {
			return PaymentChannelData{
				"channel_id":               8003200,
				"channel_item_option_info": PaymentChannelData{},
				"version":                  2,
			}
		},
	}

	IndomartISaku = PaymentChannel{
		Name: "Indomaret / i.Saku",
		ApplyFunc: func(PaymentChannelOption) PaymentChannelData {
			return PaymentChannelData{
				"channel_id": 8003001,
				"version":    2,
			}
		},
	}
)

var PaymentChannelList = [...]PaymentChannel{ShopeePay, COD, TransferBank, Alfamart, IndomartISaku}
