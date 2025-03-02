package engine

import (
	"fmt"
	"log"
	"slices"
	"time"

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
	quitChans       []chan bool
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

	log.Print("Starting job " + job.Id)

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
	s.quitChans = []chan bool{make(chan bool), make(chan bool)}

	go s.ProcessJobExecWorker()
	go s.ProcessJobInfoWorker(s.quitChans[0])
	go s.ProcessScheduleWorker(s.quitChans[1])
}

func (s *JobEngine) StopWorkers() {
	for _, ch := range s.quitChans {
		ch <- true
	}

	close(s.jobQueue)
}

func (s *JobEngine) ProcessScheduleWorker(quit chan bool) {
	lastMinute := time.Now().Minute() - 1

	for {
		select {
		case <-quit:
			return
		default:
			now := time.Now()

			if now.Minute() != lastMinute {
				lastMinute = now.Minute()
				for id, flow := range s.Flows {
					log.Printf("Checking scheduling eligibility for %s", id)
					if flow.Schedule != nil && scheduleMatches(now, flow.Schedule) {
						log.Printf("Matches!")
						s.StartJob(id)
					}
				}
			}

			time.Sleep(20 * time.Second)
		}
	}
}

func scheduleMatches(time time.Time, schedule *FlowSchedule) bool {
	return (schedule.Days == nil || slices.Contains(schedule.Days, time.Weekday())) &&
		(schedule.Hours == nil || slices.Contains(schedule.Hours, time.Hour())) &&
		(schedule.Minutes == nil || slices.Contains(schedule.Minutes, time.Minute()))
}

func (s *JobEngine) ProcessJobInfoWorker(quit chan bool) {
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
		case <-quit:
			return
		}
	}
}

func (s *JobEngine) ProcessJobExecWorker() {
	for job := range s.jobQueue {
		runJob(job, s.jobUpdates)
	}
}
