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
	pipc "github.com/nathfavour/polygeist/pkg/ipc"
	"github.com/nathfavour/polygeist/pkg/orchestrator"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	chatID := flag.String("chat", "", "Band chat room ID")
	roomID := flag.String("room", os.Getenv("BAND_CHAT_ID"), "Deprecated alias for --chat")
	apiKey := flag.String("api-key", os.Getenv("BAND_API_KEY"), "Band agent API key")
	agentID := flag.String("agent-id", os.Getenv("BAND_AGENT_ID"), "Band agent ID")
	workDir := flag.String("workdir", ".", "Default workspace directory for mutations")
	binPath := flag.String("bin", "", "Binary path for distribution phase")
	once := flag.String("once", "", "Run a single task payload locally and exit")
	showVersion := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("polygeist %s (%s)\n", version, commit)
		return
	}

	selectedChat := *chatID
	if selectedChat == "" {
		selectedChat = *roomID
	}
	if selectedChat == "" {
		selectedChat = os.Getenv("BAND_ROOM_ID")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ipcServer := pipc.NewServer()
	if err := ipcServer.Start(ctx); err != nil {
		log.Printf("ipc server disabled: %v", err)
	} else {
		log.Printf("polygeist ipc listening on %s", pipc.SocketPath())
	}

	client := band.NewAgentClient(selectedChat, *apiKey, *agentID)
	engine := orchestrator.NewSwarmEngine(client)
	engine.WorkDir = *workDir
	engine.BinPath = *binPath
	engine.IPC = ipcServer

	if *once != "" {
		ipcServer.SetPhase("mutating")
		evt := band.Event{
			EventID:  "local",
			RoomID:   selectedChat,
			Payload:  *once,
			Metadata: map[string]interface{}{"work_dir": *workDir},
		}
		if err := engine.HandleEvent(ctx, evt); err != nil {
			log.Fatalf("task failed: %v", err)
		}
		ipcServer.SetPhase("idle")
		fmt.Println("task completed")
		return
	}

	if selectedChat == "" {
		log.Fatal("chat ID required: set --chat or BAND_CHAT_ID")
	}
	if client.APIKey == "" {
		log.Fatal("API key required: set --api-key or BAND_API_KEY")
	}

	log.Printf("polygeist connecting to Band chat %s", selectedChat)
	if err := client.Listen(ctx, func(evt band.Event) error {
		ipcServer.SetPhase("running")
		defer ipcServer.SetPhase("idle")
		return engine.HandleEvent(ctx, evt)
	}); err != nil && err != context.Canceled {
		log.Fatalf("control loop stopped: %v", err)
	}
}
