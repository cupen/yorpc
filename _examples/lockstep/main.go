package main

import (
	"os"
	"syscall"

	"github.com/cupen/signalhub"
	"github.com/cupen/yorpc"
	"github.com/gin-gonic/gin"
	zapsetup "github.com/upgrade-or-die/zap-setup"
	"go.uber.org/zap"
)

var (
	log = zapsetup.RootLogger()
)

type Player struct {
	ID string
}

func (p *Player) OnCall(id uint16, data []byte) ([]byte, error) {
	return []byte{}, nil
}

func (p *Player) OnSend(id uint16, data []byte) {

}

func StartPVP() {

}

func main() {
	app := gin.Default()
	app.Static("client", "./client")
	app.GET("ws", func(c *gin.Context) {
		w, r := c.Writer, c.Request
		ws, err := yorpc.NewWithHTTP(r, w)
		if err != nil {
			panic(err)
		}
		p := Player{}
		sess := yorpc.NewRPCSession(p.ID, &p)
		sess.Connect(ws)
		if err := sess.Start(yorpc.Options{}); err != nil {
			panic(err)
		}
	})

	serverAddr := "0.0.0.0:5000"
	go func() {
		log.Info("http.listen:", zap.String("addr", serverAddr))

		if err := app.Run(serverAddr); err != nil {
			log.Fatal("http.listen failed", zap.Error(err))
		}
	}()
	h := signalhub.New()
	h.Watch(syscall.SIGTERM, func(os.Signal) {
		os.Exit(0)
	})

	log.Info("running. open http://127.0.0.1:5000/client/index.html")
	h.Run()
}
