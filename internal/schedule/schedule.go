package schedule

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

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

type JobUpdate struct {
	Id     JobId
	State  JobState
	Output string
}

type Scheduler struct {
	Jobs       map[JobId]JobState
	Flows      map[FlowId]Flow
	jobQueue   chan *Job
	jobUpdates chan JobUpdate
}

func New() Scheduler {
	return Scheduler{
		Jobs:       make(map[JobId]JobState),
		Flows:      make(map[FlowId]Flow),
		jobQueue:   make(chan *Job, 1024),
		jobUpdates: make(chan JobUpdate, 1024),
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
	s.jobUpdates <- JobUpdate{
		Id:    jobId,
		State: JobPending,
	}
	s.jobQueue <- job
	return true, jobId
}

func (s *Scheduler) ProcessUpdates() {
	for {
		select {
		case update := <-s.jobUpdates:
			s.Jobs[update.Id] = update.State
		default:
			return
		}
	}
}

func (s *Scheduler) ProcessJobsWorker() {
	for job := range s.jobQueue {
		runJob(job, func(state JobState) {
			s.jobUpdates <- JobUpdate{
				Id:    job.Id,
				State: state,
			}
		})
	}
}

func runJob(job *Job, updateState func(JobState)) bool {
	fmt.Printf("[%s] STARTED\n", job.Id)
	updateState(JobRunning)

	for i, step := range job.Steps {
		fmt.Printf("[%s] step %d\n", job.Id, i)
		if runStep(&step) != nil {
			fmt.Printf("[%s] FAILED\n", job.Id)
			updateState(JobFailed)
			return false
		}
	}

	fmt.Printf("[%s] PASSED\n", job.Id)
	updateState(JobSucceeded)
	return true
}

func runStep(step *Step) error {
	if len(step.Args) == 0 {
		return errors.New("Step is empty")
	}

	cmd := exec.Command(step.Args[0], step.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
