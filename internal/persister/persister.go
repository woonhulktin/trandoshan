package persister

import (
	"github.com/creekorful/trandoshan/internal/log"
	"github.com/creekorful/trandoshan/internal/natsutil"
	"github.com/creekorful/trandoshan/pkg/proto"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// GetApp return the persister app
func GetApp() *cli.App {
	return &cli.App{
		Name:    "trandoshan-persister",
		Version: "0.0.1",
		Usage:   "", // TODO
		Flags: []cli.Flag{
			log.GetLogFlag(),
			&cli.StringFlag{
				Name:     "nats-uri",
				Usage:    "URI to the NATS server",
				Required: true,
			},
		},
		Action: execute,
	}
}

func execute(ctx *cli.Context) error {
	log.ConfigureLogger(ctx)

	logrus.Infof("Starting trandoshan-persister v%s", ctx.App.Version)

	logrus.Debugf("Using NATS server at: %s", ctx.String("nats-uri"))

	// Create the NATS subscriber
	sub, err := natsutil.NewSubscriber(ctx.String("nats-uri"))
	if err != nil {
		return err
	}
	defer sub.Close()

	logrus.Info("Successfully initialized trandoshan-persister. Waiting for resources")

	if err := sub.QueueSubscribe(proto.ResourceSubject, "persisters", handleMessage()); err != nil {
		return err
	}

	return nil
}

func handleMessage() natsutil.MsgHandler {
	return func(nc *nats.Conn, msg *nats.Msg) error {
		var resMsg proto.ResourceMsg
		if err := natsutil.ReadJSON(msg, &resMsg); err != nil {
			return err
		}

		logrus.Debugf("Processing resource: %s", resMsg.URL)

		return nil
	}
}
