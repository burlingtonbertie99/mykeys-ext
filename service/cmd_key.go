package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/burlingtonbertie99/mykeys"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var genTypes = strings.Join([]string{
	string(keys.EdX25519),
	string(keys.X25519),
}, ", ")

func keyCommands(client *Client) []cli.Command {
	return []cli.Command{
		{
			Name:  "list",
			Usage: "List keys",
			Flags: []cli.Flag{
				cli.StringSliceFlag{Name: "type, t", Usage: "only these types (" + genTypes + ")"},
			},
			Action: func(c *cli.Context) error {
				resp, err := client.RPCClient().Keys(context.TODO(), &KeysRequest{Types: c.StringSlice("type")})
				if err != nil {
					return err
				}
				fmtKeys(resp.Keys)
				return nil
			},
		},
		{
			Name:      "key",
			Usage:     "Show key",
			ArgsUsage: "kid or user",
			Action: func(c *cli.Context) error {
				key := c.Args().First()
				if key == "" {
					return errors.Errorf("specify kid or user@service")
				}
				resp, err := client.RPCClient().Key(context.TODO(), &KeyRequest{
					Key: key,
				})
				if err != nil {
					return err
				}
				if resp.Key == nil {
					return errors.Errorf("key not found")
				}
				printMessage(resp.Key)
				return nil
			},
		},
		{
			Name:  "generate",
			Usage: "Generate a key",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "type, t", Value: "edx25519", Usage: "type (edx25519, x25519)"},
			},
			Action: func(c *cli.Context) error {
				req := &KeyGenerateRequest{
					Type: c.String("type"),
				}
				resp, err := client.RPCClient().KeyGenerate(context.TODO(), req)
				if err != nil {
					return err
				}
				fmt.Println(resp.KID)
				return nil
			},
		},
		{
			Name:  "remove",
			Usage: "Remove a key",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "kid, k", Usage: "kid"},
			},
			Action: func(c *cli.Context) error {
				kid, err := argString(c, "kid", false)
				if err != nil {
					return err
				}
				if _, err := client.RPCClient().KeyRemove(context.TODO(), &KeyRemoveRequest{
					KID: kid,
				}); err != nil {
					return err
				}
				return nil
			},
		},
	}
}
