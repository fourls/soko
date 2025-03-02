package engine

import (
	"fmt"

	"github.com/google/uuid"
)

type jobInfoRequest struct {
	id       JobId
	receiver chan *JobInfo
}

type JobEngine struct {
	jobs            map[JobId]JobInfo
	Flows           map[FlowId]Flow
	jobQueue        chan *Job
	jobUpdates      chan jobUpdate
	jobInfoRequests chan jobInfoRequest
	quit            chan int
}

func New() JobEngine {
	return JobEngine{
		jobs:            make(map[JobId]JobInfo),
		Flows:           make(map[FlowId]Flow),
		jobQueue:        make(chan *Job, 1024),
		jobUpdates:      make(chan jobUpdate, 64),
		jobInfoRequests: make(chan jobInfoRequest, 64),
	}
}

func (s *JobEngine) StartJob(flowId FlowId) (bool, JobId) {
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

func (s *JobEngine) GetJob(id JobId) *JobInfo {
	receiver := make(chan *JobInfo)
	s.jobInfoRequests <- jobInfoRequest{
		id:       id,
		receiver: receiver,
	}
	return <-receiver
}

func (s *JobEngine) StartWorkers() {
	go s.ProcessJobExecWorker()
	go s.ProcessJobInfoWorker()
}

func (s *JobEngine) StopWorkers() {
	s.quit <- 0
	s.quit <- 0
	close(s.jobQueue)
}

func (s *JobEngine) ProcessJobInfoWorker() {
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

func (s *JobEngine) ProcessJobExecWorker() {
	for job := range s.jobQueue {
		runJob(job, s.jobUpdates)
	}
}
