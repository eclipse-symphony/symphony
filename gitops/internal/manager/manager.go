package manager

import (
	"context"

	runner "dev.azure.com/msazure/One/_git/symphony/gitops/internal/runner"
)

type manager struct {
	runners map[string]runner.Runner
	ctx     context.Context
	cancel  context.CancelFunc
}

type Manager interface {
	Start()
	AddRunner(r runner.Runner)
	RemoveRunner(id string)
	Stop()
	Ctx() context.Context
}

func NewManager() *manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &manager{
		runners: make(map[string]runner.Runner),
		ctx:     ctx,
		cancel:  cancel,
	}
}
func (m *manager) Start() {
	for _, r := range m.runners {
		r.Start()
	}
}

func (m *manager) AddRunner(r runner.Runner) {
	// TODO: Rudimentary. Add lock and improve overall logic
	if _, ok := m.runners[r.GetId()]; !ok {
		m.runners[r.GetId()] = r
	} else {
		m.runners[r.GetId()].Stop()
		m.runners[r.GetId()] = r
	}
	r.Start()
}

func (m *manager) RemoveRunner(id string) {
	if r, ok := m.runners[id]; ok {
		r.Stop()
		delete(m.runners, id)
	}
}

func (m *manager) Stop() {
	for _, r := range m.runners {
		r.Stop()
	}
	m.cancel()
}

func (m *manager) Ctx() context.Context {
	return m.ctx
}
