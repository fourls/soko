package engine

type FlowId string
type JobId string

type Flow struct {
	Id       FlowId
	Steps    []Step
	Schedule *FlowSchedule
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

func (s JobState) String() string {
	switch s {
	case JobPending:
		return "pending"
	case JobRunning:
		return "running"
	case JobSucceeded:
		return "succeeded"
	case JobFailed:
		return "failed"
	default:
		return "unknown"
	}
}

type Job struct {
	Id    JobId
	Steps []Step
}

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
