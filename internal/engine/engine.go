package engine

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/fourls/soko/internal/crud"
	"github.com/google/uuid"
)

type jobInfoRequest struct {
	id       JobId
	receiver chan *JobInfo
}

type JobEngine struct {
	jobs     crud.Crud[JobId, *JobInfo]
	Flows    crud.Crud[FlowId, *Flow]
	jobQueue chan *Job
	quit     chan bool
}

// todo interfaceize this
func New() JobEngine {
	engine := JobEngine{
		jobs:     crud.New[JobId, *JobInfo](),
		Flows:    crud.New[FlowId, *Flow](),
		jobQueue: make(chan *Job, 1024),
	}
	go engine.worker()
	return engine
}

func (s *JobEngine) worker() {
	runQuit := make(chan bool)
	go s.RunJobs(runQuit)

	scheduleQuit := make(chan bool)
	go s.ProcessSchedule(scheduleQuit)

	<-s.quit
	scheduleQuit <- true
	runQuit <- true
}

func (s *JobEngine) Close() {
	s.quit <- true
}

func (s *JobEngine) StartJob(flowId FlowId) (bool, JobId) {
	jobId := JobId(fmt.Sprintf("%s:%s", flowId, uuid.New().String()))

	job := new(Job)
	job.Id = jobId

	log.Print("Starting job " + job.Id)

	flow := s.Flows.Read(flowId)
	if flow == nil {
		return false, ""
	}

	job.Steps = flow.Steps
	s.jobs.Create(jobId, &JobInfo{
		FlowId: flowId,
		Steps:  make([]StepInfo, len(flow.Steps)),
	})
	s.jobQueue <- job
	return true, jobId
}

func (s *JobEngine) GetJob(id JobId) *JobInfo {
	return s.jobs.Read(id)
}

func (s *JobEngine) ProcessSchedule(quit chan bool) {
	lastMinute := time.Now().Minute() - 1

	for {
		select {
		case <-quit:
			return
		default:
			now := time.Now()

			if now.Minute() != lastMinute {
				lastMinute = now.Minute()
				flows := s.Flows.Snapshot()
				for id, flow := range flows {
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

func (s *JobEngine) RunJobs(quit chan bool) {
	for {
		select {
		case <-quit:
			return
		case job := <-s.jobQueue:
			runJob(job, func(updateFunc func(*JobInfo)) {
				s.jobs.Update(job.Id, func(info *JobInfo) *JobInfo {
					updateFunc(info)
					return info
				})
			})
		}
	}
}
