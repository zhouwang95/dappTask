package types

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"math/big"
)

// BigInt 是*big.Int的包装类型，用于实现数据库接口
type BigInt big.Int

// Scan 实现sql.Scanner接口，用于从数据库读取值
func (b *BigInt) Scan(src interface{}) error {
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("BigInt.Scan: 不支持的类型 %T", src)
	}

	if len(data) == 0 {
		*b = BigInt(big.Int{})
		return nil
	}

	// 尝试将字符串转换为大整数
	val := new(big.Int)
	_, ok := val.SetString(string(data), 10)
	if !ok {
		return errors.New("BigInt.Scan: 无效的整数格式")
	}

	*b = BigInt(*val)
	return nil
}

// Value 实现driver.Valuer接口，用于向数据库写入值
func (b *BigInt) Value() (driver.Value, error) {
	if b == nil {
		return "0", nil
	}
	return (*big.Int)(b).String(), nil
}

// 辅助方法：转换为*big.Int
func (b *BigInt) ToBigInt() *big.Int {
	return (*big.Int)(b)
}

// 辅助方法：从*big.Int创建
func FromBigInt(val *big.Int) *BigInt {
	return (*BigInt)(val)
}
