package plugins

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

type CrashLoop struct{}

func (p *CrashLoop) Name() string { return "crash-loop" }

const crashRestartThreshold = 3

func (p *CrashLoop) Analyze(snap model.Snapshot) []model.Finding {
	type taskStat struct{ max int; worst model.Task }
	byService := make(map[string]*taskStat)

	for _, t := range snap.Tasks {
		s, ok := byService[t.ServiceID]
		if !ok {
			s = &taskStat{}
			byService[t.ServiceID] = s
		}
		if t.RestartCount > s.max {
			s.max = t.RestartCount
			s.worst = t
		}
	}

	var findings []model.Finding
	for svcID, stat := range byService {
		if stat.max < crashRestartThreshold {
			continue
		}
		svcName := stat.worst.ServiceName
		if svcName == "" {
			svcName = svcID
		}
		sev := model.SeverityMedium
		if stat.max >= 10 {
			sev = model.SeverityHigh
		}
		if stat.max >= 20 {
			sev = model.SeverityCritical
		}
		evidence := []string{fmt.Sprintf("highest restart count: %d", stat.max)}
		if stat.worst.ExitCode != 0 {
			evidence = append(evidence, fmt.Sprintf("exit code: %d", stat.worst.ExitCode))
		}
		if stat.worst.Error != "" {
			evidence = append(evidence, fmt.Sprintf("error: %s", stat.worst.Error))
		}
		findings = append(findings, model.Finding{
			ID:             uuid.NewString(),
			Severity:       sev,
			Resource:       svcName,
			Scope:          "service",
			Message:        fmt.Sprintf("Service tasks are crash-looping (%d restarts).", stat.max),
			Evidence:       evidence,
			Recommendation: "Check task logs. Common causes: bad config, missing secret, OOM kill, failed healthcheck.",
			Source:         p.Name(),
			DetectedAt:     time.Now().UTC(),
		})
	}
	return findings
}
