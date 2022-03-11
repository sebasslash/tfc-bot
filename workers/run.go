package workers

/* import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	embed "github.com/clinet/discordgo-embed"

	"github.com/hashicorp/go-tfe"
	"github.com/sebasslash/tfc-bot/models"
)

var (
	backoffMin = 1000.0
	backoffMax = 3000.0
)

type State int

const (
	Idle State = iota
	Planning
	Planned
	CostEstimating
	CostEstimated
	Applying
	Applied
)

type RunWorker struct {
	Ctx  context.Context
	step State

	ChannelID string
	RunInfo   chan *models.RunStatus
	RunAction chan *models.RunActionMessage
	RawEmbeds chan *models.EmbedMessage
	Logs      chan *models.LogMessage
	Errs      chan *models.StatusMessage

	Client        *tfe.Client
	WorkspaceID   string
	WorkspaceName string
	Organization  string
}

func (w *RunWorker) Run() error {
	err := w.findAvailableRun()
	if err != nil {
		return err
	}

	return nil
}

func (w *RunWorker) findAvailableRun() error {
	w.step = Idle
	log.Printf("[DEBUG] Finding available run on %s", w.WorkspaceID)

	wk, err := w.Client.Workspaces.ReadByIDWithOptions(w.Ctx, w.WorkspaceID, &tfe.WorkspaceReadOptions{
		Include: "current_run",
	})
	if err != nil {
		return err
	}

	if wk.CurrentRun == nil {
		return nil
	}

	prevStatus := wk.CurrentRun.Status

	for i := 0; ; i++ {
		select {
		case <-w.Ctx.Done():
			return nil
		case <-time.After(backoff(backoffMin, backoffMax, i)):
		}

		log.Printf("[DEBUG] Fetching workspace %s", w.WorkspaceID)

		// Read the workspace's current rrun
		wk, err := w.Client.Workspaces.ReadByIDWithOptions(w.Ctx, w.WorkspaceID, &tfe.WorkspaceReadOptions{
			Include: "current_run",
		})
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] Found run %s", wk.CurrentRun.ID)

		r := wk.CurrentRun

		w.setState(r.Status)

		log.Printf("[DEBUG] Run has status %s", r.Status)

		if isDiscarded(r.Status) && w.step != Idle {
			w.sendRunActionMsg(r, false)
			w.step = Idle
			continue
		}

		if isErrored(r.Status) && w.step != Idle {
			w.sendRunActionMsg(r, false)
			w.step = Idle
			continue
		}

		if isPlanned(r.Status) && w.step != Planned {
			w.step = Planned
			w.sendRunActionMsg(r, true)
			file, err := w.readPlanLogs(r)
			if err != nil {
				return err
			}

			w.Logs <- &models.LogMessage{
				Title: "Plan Output",
				File:  file,
			}
			continue
		}

		if isConfirmed(r.Status) {
			w.sendRunActionMsg(r, true)
			continue
		}

		if isCostEstimated(r.Status) && w.step != CostEstimated {
			w.step = CostEstimated
			w.runCostEstimation(r)
		}

		if isApplied(r.Status) && w.step != Applied {
			w.step = Applied
			w.sendRunActionMsg(r, true)
		}

		if prevStatus != r.Status {
			w.sendRunInfoMsg(r)
		}

		prevStatus = r.Status
	}
}

func (w *RunWorker) readPlanLogs(run *tfe.Run) (*os.File, error) {

	log.Printf("[DEBUG] Generating log file for plan %s", run.Plan.ID)

	logs, err := w.Client.Plans.Logs(w.Ctx, run.Plan.ID)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%s-plan.log", run.ID)
	logFile, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReaderSize(logs, 64*1024)
	for next := true; next; {
		var l, line []byte

		for isPrefix := true; isPrefix; {
			l, isPrefix, err = reader.ReadLine()
			if err != nil {
				if err != io.EOF {
					return nil, err
				}
				next = false
			}
			line = append(line, l...)
		}

		if next || len(line) > 0 {
			logFile.Write(line)
		}
	}

	log.Printf("[DEBUG] Done generating log file for plan %s", run.Plan.ID)

	return logFile, nil
}

func (w *RunWorker) runCostEstimation(run *tfe.Run) error {
	ce, err := w.Client.CostEstimates.Read(w.Ctx, run.CostEstimate.ID)
	if err != nil {
		return err
	}

	if ce.Status != tfe.CostEstimateFinished {
		if isErrored(run.Status) || isCanceled(run.Status) {
			return nil
		}
	}

	delta, err := strconv.ParseFloat(ce.DeltaMonthlyCost, 64)
	if err != nil {
		return err
	}

	sign := "+"
	if delta < 0 {
		sign = "-"
	}

	deltaRepr := strings.Replace(ce.DeltaMonthlyCost, "-", "", 1)
	w.sendCostEstimationMsg(run.ID, ce, sign, deltaRepr)

	return nil
}

func (w *RunWorker) sendRunInfoMsg(run *tfe.Run) {
	l := fmt.Sprintf("%s/app/%s/workspaces/%s/runs/%s",
		os.Getenv("TFE_ADDRESS"),
		w.Organization,
		w.WorkspaceName,
		run.ID)

	w.RunInfo <- &models.RunStatus{
		Title:     fmt.Sprintf("%s (workspace %s)", run.ID, w.WorkspaceName),
		Run:       run,
		ChannelID: w.ChannelID,
		PageLink:  l,
	}
}

func (w *RunWorker) sendRunActionMsg(run *tfe.Run, success bool) {
	w.RunAction <- &models.RunActionMessage{
		RunID:     run.ID,
		ChannelID: w.ChannelID,
		Status:    run.Status,
		Success:   success,
	}
}

func (w *RunWorker) sendCostEstimationMsg(runID string, ce *tfe.CostEstimate, sign string, deltaRepr string) {
	e := embed.NewEmbed().
		SetTitle("Cost Estimation for "+runID).
		AddField("Resources", fmt.Sprintf("%d of %d estimated", ce.MatchedResourcesCount, ce.ResourcesCount)).
		AddField("Cost", fmt.Sprintf("$%s/mo %s$%s", ce.ProposedMonthlyCost, sign, deltaRepr)).
		InlineAllFields().MessageEmbed

	w.RawEmbeds <- &models.EmbedMessage{
		ChannelID: w.ChannelID,
		Embed:     e,
	}
}

func (w *RunWorker) sendErrorMsg(msg, runID string) {
	w.Errs <- &models.StatusMessage{
		Title:     runID,
		Message:   msg,
		ChannelID: w.ChannelID,
	}
}

func (w *RunWorker) setState(status tfe.RunStatus) {
	if status == tfe.RunPlanQueued || status == tfe.RunPlanning {
		w.step = Planning
	}

	if status == tfe.RunCostEstimating {
		w.step = CostEstimating
	}

	if status == tfe.RunApplyQueued || status == tfe.RunApplying {
		w.step = Applying
	}
} */
