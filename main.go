package main

import (
	"flag"
	"fmt"
	"github.com/Streamlet/NoteIsSite/config"
	"github.com/Streamlet/NoteIsSite/global"
	"github.com/Streamlet/NoteIsSite/server"
	"github.com/Streamlet/NoteIsSite/util"
	"os"
	"os/signal"
	"syscall"
)

type commandLineArgs struct {
	config string
}

func main() {
	var args commandLineArgs
	flag.StringVar(&args.config, "config", "site.toml", "config file path")
	flag.Parse()

	conf, err := config.LoadSiteConfig(args.config)
	if err != nil {
		fmt.Printf("failed to load %s: %s\n", args.config, err.Error())
		return
	}

	var srv server.HttpServer
	if conf.GetSock() != "" {
		srv, err = server.NewSockServer(conf.GetSock(), conf.GetNoteDir(), conf.GetTemplateDir())
	} else if conf.GetPort() > 0 {
		srv, err = server.NewPortServer(conf.GetPort(), conf.GetNoteDir(), conf.GetTemplateDir())
	} else {
		util.Assert(false, "neither port or sock was specified")
		return
	}

	if err != nil {
		fmt.Printf("failed to init server with node dir '%s' and template dir '%s'\n", conf.GetNoteDir(), conf.GetTemplateDir())
		return
	}

	global.InitErrorChan()

	err = srv.Serve()
	if err != nil {
		fmt.Printf("failed to start server on sock '%s' or port %d\n", conf.GetSock(), conf.GetPort())
		return
	}

	fmt.Printf("Server started")
	if conf.GetSock() != "" {
		fmt.Printf(" on sock '%s'.\n", conf.GetSock())
	} else if conf.GetPort() > 0 {
		fmt.Printf(" on port %d.\n", conf.GetPort())
	}
	fmt.Printf("Note dir: %s\n", conf.GetNoteDir())
	fmt.Printf("Template dir: %s\n", conf.GetTemplateDir())
	fmt.Printf("\n")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sig:
			fmt.Printf("Server shutdown.\n")
		case err := <-global.GetErrorChan():
			fmt.Printf("Server error: %s\n", err.Error())
		default:
			continue
		}
		break
	}
}
