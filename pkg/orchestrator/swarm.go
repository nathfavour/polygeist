package orchestrator

import (
	"context"
	"fmt"

	"github.com/nathfavour/anyisland/pkg/packagemanager"
	"github.com/nathfavour/auracrab/pkg/sandbox"
	"github.com/nathfavour/polygeist/pkg/band"
	"github.com/nathfavour/vibeauracle/pkg/engine"
)

// SwarmEngine coordinates mutation, verification, and distribution phases.
type SwarmEngine struct {
	Mutator     engine.MutationInterface
	Sandbox     sandbox.ExecutionInterface
	Distributor packagemanager.DistributionInterface
	Band        *band.AgentClient
	IPC         interface{ SetPhase(string) }
	WorkDir     string
	BinPath     string
}

func NewSwarmEngine(bandClient *band.AgentClient) *SwarmEngine {
	return &SwarmEngine{
		Mutator:     engine.NewDefaultEngine(),
		Sandbox:     sandbox.NewDefaultExecutor(),
		Distributor: packagemanager.NewDefaultDistributor(""),
		Band:        bandClient,
		WorkDir:     ".",
		BinPath:     "",
	}
}

func (s *SwarmEngine) setPhase(phase string) {
	if s.IPC != nil {
		s.IPC.SetPhase(phase)
	}
}

// HandleEvent executes the three-phase Band.ai control loop for a task event.
func (s *SwarmEngine) HandleEvent(ctx context.Context, evt band.Event) error {
	workDir := evt.WorkDir()
	if s.WorkDir != "" && workDir == "." {
		workDir = s.WorkDir
	}

	s.setPhase("mutating")
	mutResult, err := s.Mutator.Mutate(ctx, engine.MutationRequest{
		WorkDir:  workDir,
		Payload:  evt.Payload,
		Metadata: evt.Metadata,
	})
	if err != nil {
		return s.fail(ctx, evt, "mutation", err)
	}
	if !mutResult.Success {
		return s.fail(ctx, evt, "mutation", fmt.Errorf("exit code %d: %s", mutResult.ExitCode, mutResult.Output))
	}

	if s.Band != nil {
		_ = s.Band.PostLog(ctx, evt.RoomID, fmt.Sprintf("[polygeist] mutation complete: %s", mutResult.Output))
	}

	// Phase 2: Verification
	s.setPhase("verifying")
	verifyResult, err := s.Sandbox.Execute(ctx, sandbox.VerifyRequest{
		WorkDir: workDir,
		Command: evt.VerifyCommand(),
	})
	if err != nil {
		return s.fail(ctx, evt, "verification", err)
	}
	if !verifyResult.Success {
		return s.fail(ctx, evt, "verification", fmt.Errorf("exit code %d: %s", verifyResult.ExitCode, verifyResult.Output))
	}

	if s.Band != nil {
		_ = s.Band.PostLog(ctx, evt.RoomID, fmt.Sprintf("[polygeist] verification passed: %s", verifyResult.Output))
	}

	// Phase 3: Distribution
	s.setPhase("distributing")
	binPath := s.BinPath
	if binPath == "" {
		binPath = workDir + "/polygeist"
	}

	distResult, err := s.Distributor.Distribute(ctx, packagemanager.DistributionRequest{
		Package: evt.PackageName(),
		Version: evt.Version(),
		WorkDir: workDir,
		BinPath: binPath,
		RoomID:  evt.RoomID,
	})
	if err != nil {
		return s.fail(ctx, evt, "distribution", err)
	}

	if s.Band != nil {
		_ = s.Band.PostRelease(ctx, evt.RoomID, distResult.Manifest)
		_ = s.Band.PostLog(ctx, evt.RoomID, fmt.Sprintf("[polygeist] release published sha256=%s url=%s", distResult.Hash, distResult.DownloadURL))
	}

	return nil
}

func (s *SwarmEngine) fail(ctx context.Context, evt band.Event, phase string, err error) error {
	if s.Band != nil {
		_ = s.Band.PostLog(ctx, evt.RoomID, fmt.Sprintf("[polygeist] %s failed: %v", phase, err))
	}
	return fmt.Errorf("%s phase: %w", phase, err)
}
