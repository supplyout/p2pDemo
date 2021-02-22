package main

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"time"
)

func main() {
	ctx := context.Background()
	a := app.New()
	ch := map[string]chan string{
		"in":  make(chan string),
		"out": make(chan string),
	}

	a.SetIcon(theme.FyneLogo())
	a.Settings().SetTheme(NewMyTheme())
	w := a.NewWindow("p2pTest")
	var (
		bootAddrStr string
		message     string
		rendezvous  string
		server      p2pServer
	)

	bootAddrBS := binding.BindString(&bootAddrStr)
	messageBS := binding.BindString(&message)
	rendezvousBS := binding.BindString(&rendezvous)
	// 将这些个binding传到p2pServer里，用于数据的更新
	bindings := make(map[string]interface{})
	bindings["bootAddrBS"] = bootAddrBS
	bindings["messageBS"] = messageBS
	bindings["rendezvousBS"] = rendezvousBS

	bootAddrEty := widget.NewEntryWithData(bootAddrBS)
	bootCt := container.NewVBox(
		widget.NewLabel("bootstrap节点地址"),
		bootAddrEty,
	)

	rendezvousEty := widget.NewEntryWithData(rendezvousBS)
	peerIDLb := widget.NewLabel("")
	peerBtAddrEty := widget.NewEntry()
	startConBtn := widget.NewButton("开始连接", func() {
		var bootstrap []string
		if s := peerBtAddrEty.Text; s != "" {
			bootstrap = append(bootstrap, s)
		}
		server = NewP2PServer(ctx, ch, bindings, SerTypeNorm, bootstrap)

	})
	peerInEty := widget.NewEntry()
	peerInBtn := widget.NewButton("发送", func() {
		if s := peerInEty.Text; s != "" {
			ch["in"] <- s + "\n"
		}
	})

	peerInEty.Hide()
	peerInBtn.Hide()
	peerIDLb.Hide()

	peerCt := container.NewVBox(
		widget.NewLabel("请输入bootstrap节点地址"),
		peerBtAddrEty,
		widget.NewLabel("请输入id"),
		rendezvousEty,
		startConBtn,
		widget.NewLabel("我的peerID"),
		peerIDLb,
		widget.NewLabel("message"),
		widget.NewLabelWithData(messageBS),
		peerInEty,
		peerInBtn,
	)

	ct := container.NewVBox(
		widget.NewButton("作为boostrap节点", func() {
			var bootstrap []string
			server = NewP2PServer(ctx, ch, bindings, SerTypeBoot, bootstrap)
			_ = bootAddrBS.Set(fmt.Sprintf("%s/p2p/%s", server.host.Addrs()[0], server.host.ID().Pretty()))
			w.SetContent(bootCt)
		}),
		widget.NewButton("作为普通节点", func() {
			w.SetContent(peerCt)
		}),
	)

	w.SetContent(ct)
	go func() {

		//ticker := time.NewTicker(time.Second * 30)

		for {
			select {
			case s := <-ch["out"]:
				//dialog.ShowInformation("", s, w)
				if s == "成功创建server" {
					time.Sleep(time.Second)
					peerBtAddrEty.Disable()
					rendezvousEty.Disable()
					startConBtn.Hide()
					peerInBtn.Show()
					peerInEty.Show()
					peerIDLb.SetText(server.host.ID().Pretty())
					peerIDLb.Show()
					_ = messageBS.Set(s)
				}
				_ = messageBS.Set(s)
				//case <-ticker.C:
				//	fmt.Printf("tick\n")
			}
		}
	}()

	w.ShowAndRun()

}
