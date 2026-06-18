package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nathfavour/polygeist/pkg/band"
	"github.com/nathfavour/polygeist/pkg/orchestrator"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	roomID := flag.String("room", os.Getenv("BAND_ROOM_ID"), "Band.ai room ID")
	token := flag.String("token", os.Getenv("BAND_TOKEN"), "Band.ai API token")
	workDir := flag.String("workdir", ".", "Default workspace directory for mutations")
	binPath := flag.String("bin", "", "Binary path for distribution phase")
	once := flag.String("once", "", "Run a single task payload locally and exit")
	showVersion := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("polygeist %s (%s)\n", version, commit)
		return
	}

	client := band.NewClient(*roomID, *token)
	engine := orchestrator.NewSwarmEngine(client)
	engine.WorkDir = *workDir
	engine.BinPath = *binPath

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if *once != "" {
		evt := band.Event{
			EventID:  "local",
			RoomID:   *roomID,
			Payload:  *once,
			Metadata: map[string]interface{}{"work_dir": *workDir},
		}
		if err := engine.HandleEvent(ctx, evt); err != nil {
			log.Fatalf("task failed: %v", err)
		}
		fmt.Println("task completed")
		return
	}

	if *roomID == "" {
		log.Fatal("room ID required: set --room or BAND_ROOM_ID")
	}

	log.Printf("polygeist listening on room %s", *roomID)
	err := client.Listen(ctx, func(evt band.Event) error {
		log.Printf("event %s from %s", evt.EventID, evt.SenderID)
		return engine.HandleEvent(ctx, evt)
	})
	if err != nil && err != context.Canceled {
		log.Fatalf("control loop stopped: %v", err)
	}
}
