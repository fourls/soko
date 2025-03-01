package schedule

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/uuid"
)

type FlowId string
type JobId string

type Flow struct {
	Id    FlowId
	Steps []Step
}

type Step struct {
	Args []string
}

type JobState int

const (
	JobPending JobState = iota
	JobRunning
	JobSucceeded
	JobFailed
)

type Job struct {
	Id    JobId
	Steps []Step
}

type jobUpdate interface {
	JobId() JobId
}

type jobMetaUpdate interface {
	JobId() JobId
	FlowId() FlowId
}

type jobStateUpdate interface {
	JobId() JobId
	JobState() JobState
}

type jobStepUpdate interface {
	JobId() JobId
	StepIndex() int
	StepInput() string
	StepOutput() []byte
}

type jobInfoInit struct {
	id     JobId
	flowId FlowId
}

func (j jobInfoInit) JobId() JobId       { return j.id }
func (j jobInfoInit) FlowId() FlowId     { return j.flowId }
func (j jobInfoInit) JobState() JobState { return JobPending }

type miscJobStateUpdate struct {
	id       JobId
	jobState JobState
}

func (j miscJobStateUpdate) JobId() JobId       { return j.id }
func (j miscJobStateUpdate) JobState() JobState { return j.jobState }

type jobStepUpdateImpl struct {
	id         JobId
	stepIndex  int
	jobState   JobState
	stepInput  string
	stepOutput []byte
}

func (j jobStepUpdateImpl) JobId() JobId       { return j.id }
func (j jobStepUpdateImpl) JobState() JobState { return j.jobState }
func (j jobStepUpdateImpl) StepIndex() int     { return j.stepIndex }
func (j jobStepUpdateImpl) StepInput() string  { return j.stepInput }
func (j jobStepUpdateImpl) StepOutput() []byte { return j.stepOutput }

var (
	_ jobStepUpdate  = jobStepUpdateImpl{}
	_ jobStateUpdate = jobStepUpdateImpl{}
)

type JobInfo struct {
	FlowId      FlowId
	State       JobState
	CurrentStep int
	Steps       []StepInfo
}

type StepInfo struct {
	Input  string
	Output []byte
}

type jobInfoRequest struct {
	id       JobId
	receiver chan *JobInfo
}

type Scheduler struct {
	jobs            map[JobId]JobInfo
	Flows           map[FlowId]Flow
	jobQueue        chan *Job
	jobUpdates      chan jobUpdate
	jobInfoRequests chan jobInfoRequest
	quit            chan int
}

func New() Scheduler {
	return Scheduler{
		jobs:            make(map[JobId]JobInfo),
		Flows:           make(map[FlowId]Flow),
		jobQueue:        make(chan *Job, 1024),
		jobUpdates:      make(chan jobUpdate, 64),
		jobInfoRequests: make(chan jobInfoRequest, 64),
	}
}

func (s *Scheduler) StartJob(flowId FlowId) (bool, JobId) {
	jobId := JobId(fmt.Sprintf("%s:%s", flowId, uuid.New().String()))

	job := new(Job)
	job.Id = jobId

	flow, ok := s.Flows[flowId]
	if !ok {
		return false, ""
	}

	job.Steps = flow.Steps
	s.jobUpdates <- jobInfoInit{id: jobId, flowId: flowId}
	s.jobQueue <- job
	return true, jobId
}

func (s *Scheduler) GetJob(id JobId) *JobInfo {
	receiver := make(chan *JobInfo)
	s.jobInfoRequests <- jobInfoRequest{
		id:       id,
		receiver: receiver,
	}
	return <-receiver
}

func (s *Scheduler) StartWorkers() {
	go s.ProcessJobExecWorker()
	go s.ProcessJobInfoWorker()
}

func (s *Scheduler) StopWorkers() {
	s.quit <- 0
}

func (s *Scheduler) ProcessJobInfoWorker() {
	for {
		select {
		case request := <-s.jobInfoRequests:
			info, ok := s.jobs[request.id]
			if ok {
				request.receiver <- &info
			} else {
				request.receiver <- nil
			}
		case update := <-s.jobUpdates:
			info, ok := s.jobs[update.JobId()]
			if !ok {
				info = JobInfo{Steps: make([]StepInfo, 0)}
			}

			if metaUpdate, ok := update.(jobMetaUpdate); ok {
				flowId := metaUpdate.FlowId()

				info = JobInfo{
					FlowId: metaUpdate.FlowId(),
					Steps:  make([]StepInfo, len(s.Flows[flowId].Steps)),
				}
			}

			if stateUpdate, ok := update.(jobStateUpdate); ok {
				info.State = stateUpdate.JobState()
			}

			if stepUpdate, ok := update.(jobStepUpdate); ok {
				i := stepUpdate.StepIndex()
				if i >= len(info.Steps) {
					info.Steps = append(info.Steps, StepInfo{})
				}

				info.CurrentStep = i
				info.Steps[i] = StepInfo{
					Input:  stepUpdate.StepInput(),
					Output: stepUpdate.StepOutput(),
				}
			}

			s.jobs[update.JobId()] = info
		case <-s.quit:
			return
		}
	}
}

func (s *Scheduler) ProcessJobExecWorker() {
	for job := range s.jobQueue {
		runJob(job, s.jobUpdates)
	}
}

func runJob(job *Job, updateChan chan jobUpdate) bool {
	updateChan <- miscJobStateUpdate{id: job.Id, jobState: JobRunning}

	for i, step := range job.Steps {
		var output []byte = nil
		var state JobState = JobRunning

		output, err := runStep(&step)
		if err != nil {
			state = JobFailed
			output = []byte(fmt.Sprintf("Step failed with error:\n  %s\n\n%s", err.Error(), output))
		}

		updateChan <- jobStepUpdateImpl{
			id:         job.Id,
			stepIndex:  i,
			jobState:   state,
			stepInput:  strings.Join(step.Args, " "),
			stepOutput: output,
		}

		if state == JobFailed {
			return false
		}
	}

	updateChan <- miscJobStateUpdate{id: job.Id, jobState: JobSucceeded}
	return true
}

func runStep(step *Step) ([]byte, error) {
	if len(step.Args) == 0 {
		return nil, errors.New("Step is empty")
	}

	cmd := exec.Command(step.Args[0], step.Args[1:]...)
	return cmd.CombinedOutput()
}
