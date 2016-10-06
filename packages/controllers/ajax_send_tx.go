// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package controllers

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/lib"
	"github.com/DayLightProject/go-daylight/packages/script"
	"github.com/DayLightProject/go-daylight/packages/smart"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

const ASendTx = `ajax_send_tx`

type SendTxJson struct {
	Error string `json:"error"`
}

func init() {
	newPage(ASendTx, `json`)
}

func (c *Controller) AjaxSendTx() interface{} {
	var (
		result SendTxJson
		err    error
	)
	fmt.Println(`TXSend 0`)
	cntname := c.r.FormValue(`TxName`)
	contract := smart.GetContract(cntname)
	if contract == nil || contract.Block.Info.(*script.ContractInfo).Tx == nil {
		err = fmt.Errorf(`there is not %s contract`, cntname)
	} else {
		//		info := (*contract).Block.Info.(*script.ContractInfo)

		userId := c.SessWalletId
		if c.SessStateId > 0 {
			userId = c.SessCitizenId
		}

		/*		forsign := fmt.Sprintf("%d,%d,%d,%d", info.Id, c.r.FormValue(`time`), userId, c.SessStateId)

				for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
					val := c.r.FormValue(fitem.Name)
					if len(val) == 0 && !strings.Contains(fitem.Tags, `optional`) {
						err = fmt.Errorf(`%s is empty`, fitem.Name)
						break
					}
					forsign += fmt.Sprintf(",%v", val)
				}*/
		sign := make([]byte, 0)
		for i := 1; i <= 3; i++ {
			signature := utils.ConvertJSSign(c.r.FormValue(fmt.Sprintf("signature%d", i)))
			if i == 1 || len(signature) > 0 {
				bsign, _ := hex.DecodeString(signature)
				//				sign = append(sign, utils.EncodeLengthPlusData(bsign)...)
				lib.EncodeLenByte(&sign, bsign)
			}
		}
		if len(sign) == 0 {
			result.Error = `signature is empty`
		} else {
			//			var (
			data := make([]byte, 0)
			//			)
			header := consts.TXHeader{
				Type:    int32(contract.Block.Info.(*script.ContractInfo).Id), /* + smart.CNTOFF*/
				Time:    uint32(utils.StrToInt64(c.r.FormValue(`time`))),
				UserId:  userId,
				StateId: int64(c.SessStateId),
				Sign:    sign,
			}
			fmt.Println(`SEND TX`, contract.Block.Info.(*script.ContractInfo))
			_, err = lib.BinMarshal(&data, &header)
			if err == nil {
			fields:
				for _, fitem := range *contract.Block.Info.(*script.ContractInfo).Tx {
					val := c.r.FormValue(fitem.Name)
					if len(val) == 0 && !strings.Contains(fitem.Tags, `optional`) {
						err = fmt.Errorf(`%s is empty`, fitem.Name)
						break
					}
					fmt.Println(fitem, fitem.Type.Name(), fitem.Type.String())
					switch fitem.Type.String() {
					case `int64`:
						lib.EncodeLenInt64(&data, utils.StrToInt64(val))
					case `string`:
						data = append(append(data, lib.EncodeLength(int64(len(val)))...), []byte(val)...)
					case `[]uint8`:
						var bytes []byte
						bytes, err = hex.DecodeString(val)
						if err != nil {
							break fields
						}
						data = append(append(data, lib.EncodeLength(int64(len(bytes)))...), bytes...)
					}
				}
				if err == nil {
					md5 := utils.Md5(data)
					err = c.ExecSql(`INSERT INTO transactions_status (
						hash, time,	type, wallet_id, citizen_id	) VALUES (
						[hex], ?, ?, ?, ? )`, md5, time.Now().Unix(), header.Type, userId, c.SessStateId)
					if err == nil {
						err = c.ExecSql("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", md5, hex.EncodeToString(data))
					}
				}
			}
			fmt.Printf("Data %v %d %x", err, len(data), data)
		}
	}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}
