package ui

import (
	"fmt"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"github.com/decred/dcrd/dcrutil"
	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
	"github.com/planetdecred/godcr/wallet"
)

const PageUTXO = "unspentTransactionOutput"

type utxoPage struct {
	utxoListContainer      layout.List
	txAuthor               *dcrlibwallet.TxAuthor
	line                   *decredmaterial.Line
	backButton             decredmaterial.IconButton
	useUTXOButton          decredmaterial.Button
	unspentOutputs         **wallet.UnspentOutputs
	unspentOutputsSelected *map[int]map[int32]map[string]*wallet.UnspentOutput
	checkboxes             []decredmaterial.CheckBoxStyle
	selecAllChexBox        decredmaterial.CheckBoxStyle

	txnFee            string
	txnAmount         string
	txnAmountAfterFee string

	selectedWalletID  int
	selectedAccountID int32
}

func (win *Window) UTXOPage(common pageCommon) layout.Widget {
	pg := &utxoPage{
		unspentOutputs: &win.walletUnspentOutputs,
		utxoListContainer: layout.List{
			Axis: layout.Vertical,
		},
		line:                   common.theme.Line(),
		txAuthor:               &win.txAuthor,
		unspentOutputsSelected: &common.selectedUTXO,
		selecAllChexBox:        common.theme.CheckBox(new(widget.Bool), ""),
	}
	pg.line.Color = common.theme.Color.Gray
	pg.line.Height = 1

	pg.backButton = common.theme.PlainIconButton(new(widget.Clickable), common.icons.navigationArrowBack)
	pg.backButton.Color = common.theme.Color.Hint
	pg.backButton.Size = values.MarginPadding30
	pg.useUTXOButton = common.theme.Button(new(widget.Clickable), "OK")

	return func(gtx C) D {
		pg.Handler(common)
		return pg.Layout(gtx, common)
	}
}

func (pg *utxoPage) Handler(common pageCommon) {
	pg.selectedWalletID = common.info.Wallets[*common.selectedWallet].ID
	pg.selectedAccountID = common.info.Wallets[*common.selectedWallet].Accounts[*common.selectedAccount].Number

	if len(pg.checkboxes) != len((*pg.unspentOutputs).List) {
		pg.checkboxes = make([]decredmaterial.CheckBoxStyle, len((*pg.unspentOutputs).List))
		for i := 0; i < len((*pg.unspentOutputs).List); i++ {
			utxo := (*pg.unspentOutputs).List[i]
			pg.checkboxes[i] = common.theme.CheckBox(new(widget.Bool), "")
			if _, ok := (*pg.unspentOutputsSelected)[pg.selectedWalletID][pg.selectedAccountID][utxo.UTXO.OutputKey]; ok {
				pg.checkboxes[i].CheckBox.Value = true
			}
		}
		pg.calculateAmountAndFeeUTXO()
	}

	if pg.backButton.Button.Clicked() {
		pg.clearPageData()
		*common.page = PageSend
	}

	if pg.useUTXOButton.Button.Clicked() {
		*common.page = PageSend
	}

	if pg.selecAllChexBox.CheckBox.Changed() {
		for i, utxo := range (*pg.unspentOutputs).List {
			if pg.selecAllChexBox.CheckBox.Value {
				pg.checkboxes[i].CheckBox.Value = true
				(*pg.unspentOutputsSelected)[pg.selectedWalletID][pg.selectedAccountID][utxo.UTXO.OutputKey] = utxo
			} else {
				delete((*pg.unspentOutputsSelected)[pg.selectedWalletID][pg.selectedAccountID], utxo.UTXO.OutputKey)
				pg.checkboxes[i].CheckBox.Value = false
			}
		}
		pg.calculateAmountAndFeeUTXO()
	}
}

func (pg *utxoPage) handlerCheckboxes(cb *decredmaterial.CheckBoxStyle, utxo *wallet.UnspentOutput) {
	if cb.CheckBox.Changed() {
		if cb.CheckBox.Value {
			(*pg.unspentOutputsSelected)[pg.selectedWalletID][pg.selectedAccountID][utxo.UTXO.OutputKey] = utxo
		} else {
			delete((*pg.unspentOutputsSelected)[pg.selectedWalletID][pg.selectedAccountID], utxo.UTXO.OutputKey)
		}
		pg.calculateAmountAndFeeUTXO()
	}
}

func (pg *utxoPage) calculateAmountAndFeeUTXO() {
	var utxoKeys []string
	var totalAmount int64
	for utxoKey, utxo := range (*pg.unspentOutputsSelected)[pg.selectedWalletID][pg.selectedAccountID] {
		utxoKeys = append(utxoKeys, utxoKey)
		totalAmount += utxo.UTXO.Amount
	}
	err := pg.txAuthor.UseInputs(utxoKeys)
	if err != nil {
		log.Error(err)
		return
	}
	feeAndSize, err := pg.txAuthor.EstimateFeeAndSize()
	if err != nil {
		log.Error(err)
		return
	}
	pg.txnAmount = dcrutil.Amount(totalAmount).String()
	pg.txnFee = dcrutil.Amount(feeAndSize.Fee.AtomValue).String()
	pg.txnAmountAfterFee = dcrutil.Amount(totalAmount - feeAndSize.Fee.AtomValue).String()
}

func (pg *utxoPage) clearPageData() {
	pg.checkboxes = nil
	pg.txnFee = ""
}

func (pg *utxoPage) Layout(gtx layout.Context, c pageCommon) layout.Dimensions {
	return c.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.W.Layout(gtx, func(gtx C) D { return pg.backButton.Layout(gtx) })
					}),
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Left: values.MarginPadding10, Top: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
							return c.theme.H5("Coin Control").Layout(gtx)
						})
					}),
				)
			}),
			layout.Flexed(1, func(gtx C) D {
				return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							return layout.Inset{Bottom: values.MarginPadding15}.Layout(gtx, func(gtx C) D {
								return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										utxos := (*pg.unspentOutputsSelected)[pg.selectedWalletID][pg.selectedAccountID]
										return textData(gtx, &c, "Quantity:  ", fmt.Sprintf("%d", len(utxos)))
									}),
									layout.Rigid(func(gtx C) D {
										return textData(gtx, &c, "Amount:  ", pg.txnAmount)
									}),
									layout.Rigid(func(gtx C) D {
										return textData(gtx, &c, "Fee:  ", pg.txnFee)
									}),
									layout.Rigid(func(gtx C) D {
										return textData(gtx, &c, "After Fee:  ", pg.txnAmountAfterFee)
									}),
								)
							})
						}),
						layout.Rigid(func(gtx C) D {
							pg.line.Width = gtx.Constraints.Max.X
							return pg.line.Layout(gtx)
						}),
						layout.Rigid(func(gtx C) D {
							return pg.utxoRowHeader(gtx, &c)
						}),
						layout.Flexed(1, func(gtx C) D {
							if len(pg.checkboxes) == 0 {
								return layout.Dimensions{}
							}
							return pg.utxoListContainer.Layout(gtx, len((*pg.unspentOutputs).List), func(gtx C, index int) D {
								utxo := (*pg.unspentOutputs).List[index]
								pg.handlerCheckboxes(&pg.checkboxes[index], utxo)
								return pg.utxoRow(gtx, utxo, &c, index)
							})
						}),
						layout.Rigid(func(gtx C) D {
							return pg.useUTXOButton.Layout(gtx)
						}),
					)
				})
			}),
		)
	})
}

func textData(gtx layout.Context, c *pageCommon, txt, value string) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(c.theme.Label(values.MarginPadding15, txt).Layout),
		layout.Rigid(c.theme.Label(values.MarginPadding15, value).Layout),
	)
}

func (pg *utxoPage) utxoRowHeader(gtx layout.Context, c *pageCommon) layout.Dimensions {
	txt := c.theme.Label(values.MarginPadding15, "")
	txt.MaxLines = 1
	return layout.Inset{Top: values.MarginPadding10, Bottom: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Px(values.MarginPadding35)
				return pg.selecAllChexBox.Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Px(values.MarginPadding150)
				txt.Text = "Amount"
				return txt.Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Px(values.MarginPadding200)
				txt.Text = "Address"
				return txt.Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Px(values.MarginPadding100)
				txt.Text = "Date (UTC)"
				txt.Alignment = text.End
				return txt.Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Px(values.MarginPadding100)
				txt.Text = "Confirmations"
				return txt.Layout(gtx)
			}),
		)
	})
}

func (pg *utxoPage) utxoRow(gtx layout.Context, data *wallet.UnspentOutput, c *pageCommon, index int) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return pg.checkboxes[index].Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			txt := c.theme.Body2(data.Amount)
			txt.MaxLines = 1
			txt.Alignment = text.Start
			gtx.Constraints.Min.X = gtx.Px(values.MarginPadding150)
			return txt.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			txt := c.theme.Body2(data.UTXO.Addresses)
			txt.MaxLines = 1
			gtx.Constraints.Max.X = gtx.Px(values.MarginPadding200)
			gtx.Constraints.Min.X = gtx.Px(values.MarginPadding200)
			return txt.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			txt := c.theme.Body2(data.DateTime)
			txt.MaxLines = 1
			txt.Alignment = text.End
			gtx.Constraints.Min.X = gtx.Px(values.MarginPadding100)
			return txt.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			txt := c.theme.Body2(fmt.Sprintf("%d", data.UTXO.Confirmations))
			txt.MaxLines = 1
			txt.Alignment = text.End
			gtx.Constraints.Min.X = gtx.Px(values.MarginPadding100)
			return txt.Layout(gtx)
		}),
	)
}