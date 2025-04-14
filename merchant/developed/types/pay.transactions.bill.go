package types

import (
	"net/url"
	"time"
)

// TradeBillQuest 申请交易账单
type TradeBillQuest struct {
	//【账单日期】 账单日期，格式yyyy-MM-DD，仅支持三个月内的账单下载申请。
	BillDate string `json:"bill_date"`
	//【账单类型】 账单类型，不填则默认是ALL
	//	可选取值
	//	ALL: 返回当日所有订单信息（不含充值退款订单）
	//	SUCCESS: 返回当日成功支付的订单（不含充值退款订单）
	//	REFUND: 返回当日退款订单（不含充值退款订单）
	BillType string `json:"bill_type,omitempty"`
	//【压缩类型】 压缩类型，不填则以不压缩的方式返回账单文件流。
	//	可选取值：
	//	GZIP: 下载账单时返回.gzip格式的压缩文件流
	TarType string `json:"tar_type,omitempty"`
}

func (q *TradeBillQuest) ToUrlValues() url.Values {
	values := url.Values{}

	// 账单日期是必填字段，直接添加
	if q.BillDate == "" {
		q.BillDate = time.Now().Format("2006-01-02")
	}
	values.Add("bill_date", q.BillDate)

	// 账单类型是可选字段，只有在非空时才添加
	if q.BillType != "" {
		values.Add("bill_type", q.BillType)
	}

	// 压缩类型是可选字段，只有在非空时才添加
	if q.TarType != "" {
		values.Add("tar_type", q.TarType)
	}

	return values
}

// BillResp 应答参数
type BillResp struct {
	//【哈希类型】 哈希类型，固定为SHA1。
	HashType string `json:"hash_type"`
	//【哈希值】 账单文件的SHA1摘要值，用于商户侧校验文件的一致性。
	HashValue string `json:"hash_value"`
	//【下载地址】 供下一步请求账单文件的下载地址，该地址5min内有效。
	//	参考下载账单
	DownloadUrl string `json:"download_url"`
}

// FundsBillQuest 申请资金账单
type FundsBillQuest struct {
	//【账单日期】 账单日期，格式yyyy-MM-DD，仅支持三个月内的账单下载申请。
	BillDate string `json:"bill_date"`
	//【资金账户类型】 资金账户类型，不填默认是BASIC
	//	可选取值
	//	BASIC: 基本账户
	//	OPERATION: 运营账户
	//	FEES: 手续费账户
	AccountType string `json:"account_type,omitempty"`
	//【压缩类型】 压缩类型，不填则以不压缩的方式返回账单文件流。
	//	可选取值：
	//	GZIP: 下载账单时返回.gzip格式的压缩文件流
	TarType string `json:"tar_type,omitempty"`
}

func (q *FundsBillQuest) ToUrlValues() url.Values {
	values := url.Values{}

	// 账单日期是必填字段，直接添加
	if q.BillDate == "" {
		q.BillDate = time.Now().Format("2006-01-02")
	}
	values.Add("bill_date", q.BillDate)

	// 账单类型是可选字段，只有在非空时才添加
	if q.AccountType != "" {
		values.Add("account_type", q.AccountType)
	}

	// 压缩类型是可选字段，只有在非空时才添加
	if q.TarType != "" {
		values.Add("tar_type", q.TarType)
	}

	return values
}
