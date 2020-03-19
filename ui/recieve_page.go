package ui

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"github.com/atotto/clipboard"
	"github.com/raedahgroup/godcr-gio/ui/decredmaterial"
	"github.com/skip2/go-qrcode"
)

var (
	listContainer = &layout.List{Axis: layout.Vertical}
)

func (win *Window) Receive() {
	if win.walletInfo.LoadedWallets == 0 {
		win.Page(func() {
			win.outputs.noWallet.Layout(win.gtx)
		})
		return
	}

	info := win.walletInfo.Wallets[win.selected]
 	for i := range info.Accounts {
		if len(info.Accounts) == 0 {
			continue
		}
		
		account := info.Accounts[i]

		win.outputs.selectedWalletNameLabel.Text = info.Name
		win.outputs.selectedWalletBalLabel.Text = info.Balance
		win.outputs.selectedAccountNameLabel.Text = account.Name
		win.outputs.selectedAccountBalanceLabel.Text = account.SpendableBalance
		win.addrs = account.CurrentAddress
		break
	}

	body := func() {
		layout.Stack{}.Layout(win.gtx,
			layout.Expanded(func() {
				win.ReceivePageContents()
			}),
		)
	}
	win.TabbedPage(body)
}

func (win *Window) ReceivePageContents() {
	ReceivePageContent := []func(){
		func() {
			win.pageFirstColumn()
		},
		func() {
			layout.Center.Layout(win.gtx, func() {
				if win.outputs.err.Text != "" {
					win.outputs.err.Layout(win.gtx)
				}
			})
		},
		func() {
			win.selectedAccountColumn()
		},
		func() {
			win.qrCodeAddressColumn()
		},
		func() {
			layout.Flex{}.Layout(win.gtx,
				layout.Flexed(.35, func() {
				}),
				layout.Flexed(1, func() {
					if win.addressCopiedLabel.Text != "" {
						win.addressCopiedLabel.Layout(win.gtx)
					}
				}),
			)
		},
	}

	listContainer.Layout(win.gtx, len(ReceivePageContent), func(i int) {
		layout.Inset{Left: unit.Dp(20)}.Layout(win.gtx, ReceivePageContent[i])
	})
}

func (win *Window) pageFirstColumn() {
	layout.Flex{}.Layout(win.gtx,
		layout.Flexed(.6, func() {
			win.outputs.pageTitle.Layout(win.gtx)
		}),
		layout.Flexed(.4, func() {
			layout.Inset{Top: unit.Dp(4), Bottom: unit.Dp(20)}.Layout(win.gtx, func() {
				layout.Flex{}.Layout(win.gtx,
					layout.Flexed(.5, func() {
						win.outputs.info.Layout(win.gtx, &win.inputs.receiveIcons.info)
					}),
					layout.Flexed(.5, func() {
						win.outputs.more.Layout(win.gtx, &win.inputs.receiveIcons.more)
					}),
				)
			})
		}),
	)
}

func (win *Window) selectedAccountColumn() {
	layout.Flex{}.Layout(win.gtx,
		layout.Flexed(0.22, func() {
		}),
		layout.Flexed(1, func() {
			layout.Stack{}.Layout(win.gtx,
				layout.Stacked(func() {
					selectedDetails := func() {
						layout.UniformInset(unit.Dp(10)).Layout(win.gtx, func() {
							layout.Flex{Axis: layout.Vertical}.Layout(win.gtx,
								layout.Rigid(func() {
									layout.Flex{}.Layout(win.gtx,
										layout.Rigid(func() {
											layout.Inset{Bottom: unit.Dp(5)}.Layout(win.gtx, func() {
												win.outputs.selectedAccountNameLabel.Layout(win.gtx)
											})
										}),
										layout.Rigid(func() {
											layout.Inset{Left: unit.Dp(20)}.Layout(win.gtx, func() {
												win.outputs.selectedAccountBalanceLabel.Layout(win.gtx)
											})
										}),
									)
								}),
								layout.Rigid(func() {
									layout.Inset{Left: unit.Dp(20)}.Layout(win.gtx, func() {
										layout.Flex{}.Layout(win.gtx,
											layout.Rigid(func() {
												layout.Inset{Bottom: unit.Dp(5)}.Layout(win.gtx, func() {
													win.outputs.selectedWalletNameLabel.Layout(win.gtx)
												})
											}),
											layout.Rigid(func() {
												layout.Inset{Left: unit.Dp(22)}.Layout(win.gtx, func() {
													win.outputs.selectedWalletBalLabel.Layout(win.gtx)
												})
											}),
										)
									})
								}),
							)
						})
					}
					decredmaterial.Card{}.Layout(win.gtx, selectedDetails)
				}),
			)
		}),
	)
}

func (win *Window) qrCodeAddressColumn() {
	qrCode, err := qrcode.New(win.addrs, qrcode.Highest)
	if err != nil {
		win.outputs.err.Text = err.Error()
		return
	}

	qrCode.DisableBorder = true
	layout.Flex{Axis: layout.Vertical}.Layout(win.gtx,
		layout.Rigid(func() {
			layout.Inset{Top: unit.Dp(16)}.Layout(win.gtx, func() {
				layout.Flex{}.Layout(win.gtx,
					layout.Flexed(0.2, func() {
					}),
					layout.Flexed(1, func() {
						img := win.theme.Image(paint.NewImageOp(qrCode.Image(520)))
						img.Src.Rect.Max.X = 521
						img.Src.Rect.Max.Y = 521
						img.Scale = 0.5
						img.Layout(win.gtx)
					}),
				)
			})
		}),
		layout.Rigid(func() {
			layout.Inset{Top: unit.Dp(16)}.Layout(win.gtx, func() {
				layout.Flex{}.Layout(win.gtx,
					layout.Flexed(0.1, func() {
					}),
					layout.Flexed(1, func() {
						win.receiveAddressColumn()
					}),
				)
			})
		}),
	)
}

func (win *Window) receiveAddressColumn() {
	layout.Flex{}.Layout(win.gtx,
		layout.Rigid(func() {
			win.outputs.receiveAddressLabel.Text = win.addrs
			win.outputs.receiveAddressLabel.Layout(win.gtx)
		}),
		layout.Rigid(func() {
			layout.Inset{Left: unit.Dp(16)}.Layout(win.gtx, func() {
				for win.inputs.receiveIcons.copy.Clicked(win.gtx) {
					clipboard.WriteAll(win.addrs)
					win.addressCopiedLabel.Text = "Address Copied"
					time.AfterFunc(time.Second*9, func() {
						win.addressCopiedLabel.Text = ""
					})
				}
				win.outputs.copy.Layout(win.gtx, &win.inputs.receiveIcons.copy)
			})
		}),
	)
}
