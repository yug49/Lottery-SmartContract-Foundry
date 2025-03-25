package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/pelletier/go-toml/v2"
	"google.golang.org/grpc"

	jobv1 "github.com/smartcontractkit/chainlink-protos/job-distributor/v1/job"
	"github.com/smartcontractkit/chainlink/v2/core/services/feeds"
	"github.com/smartcontractkit/chainlink/v2/core/services/job"
)

type JobServiceClient struct {
	jobStore
	proposalStore
	nodeStore
}

func NewJobServiceClient(ns nodeStore) *JobServiceClient {
	return &JobServiceClient{
		jobStore:      newMapJobStore(),
		proposalStore: newMapProposalStore(),
		nodeStore:     ns,
	}
}

func (j *JobServiceClient) BatchProposeJob(ctx context.Context, in *jobv1.BatchProposeJobRequest, opts ...grpc.CallOption) (*jobv1.BatchProposeJobResponse, error) {
	targets := make(map[string]Node)
	for _, nodeID := range in.NodeIds {
		node, err := j.nodeStore.get(nodeID)
		if err != nil {
			return nil, fmt.Errorf("node not found: %s", nodeID)
		}
		targets[nodeID] = *node
	}
	if len(targets) == 0 {
		return nil, errors.New("no nodes found")
	}
	out := &jobv1.BatchProposeJobResponse{
		SuccessResponses: make(map[string]*jobv1.ProposeJobResponse),
		FailedResponses:  make(map[string]*jobv1.ProposeJobFailure),
	}
	var totalErr error
	for id := range targets {
		singleReq := &jobv1.ProposeJobRequest{
			NodeId: id,
			Spec:   in.Spec,
			Labels: in.Labels,
		}
		resp, err := j.ProposeJob(ctx, singleReq)
		if err != nil {
			out.FailedResponses[id] = &jobv1.ProposeJobFailure{
				ErrorMessage: err.Error(),
			}
			totalErr = errors.Join(totalErr, fmt.Errorf("failed to propose job for node %s: %w", id, err))
		}
		out.SuccessResponses[id] = resp
	}
	return out, totalErr
}

func (j *JobServiceClient) UpdateJob(ctx context.Context, in *jobv1.UpdateJobRequest, opts ...grpc.CallOption) (*jobv1.UpdateJobResponse, error) {
	// TODO CCIP-3108 implement me
	panic("implement me")
}

func (j *JobServiceClient) GetJob(ctx context.Context, in *jobv1.GetJobRequest, opts ...grpc.CallOption) (*jobv1.GetJobResponse, error) {
	// implementation detail that job id and uuid is the same
	jb, err := j.jobStore.get(in.GetId())
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	// TODO CCIP-3108 implement me
	return &jobv1.GetJobResponse{
		Job: jb,
	}, nil
}

func (j *JobServiceClient) GetProposal(ctx context.Context, in *jobv1.GetProposalRequest, opts ...grpc.CallOption) (*jobv1.GetProposalResponse, error) {
	p, err := j.proposalStore.get(in.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}
	return &jobv1.GetProposalResponse{
		Proposal: p,
	}, nil
}

func (j *JobServiceClient) ListJobs(ctx context.Context, in *jobv1.ListJobsRequest, opts ...grpc.CallOption) (*jobv1.ListJobsResponse, error) {
	jbs, err := j.jobStore.list(in.Filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	return &jobv1.ListJobsResponse{
		Jobs: jbs,
	}, nil
}

func (j *JobServiceClient) ListProposals(ctx context.Context, in *jobv1.ListProposalsRequest, opts ...grpc.CallOption) (*jobv1.ListProposalsResponse, error) {
	proposals, err := j.proposalStore.list(in.Filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list proposals: %w", err)
	}
	return &jobv1.ListProposalsResponse{
		Proposals: proposals,
	}, nil
}

// ProposeJob is used to propose a job to the node
// It auto approves the job
func (j *JobServiceClient) ProposeJob(ctx context.Context, in *jobv1.ProposeJobRequest, opts ...grpc.CallOption) (*jobv1.ProposeJobResponse, error) {
	n, err := j.nodeStore.get(in.NodeId)
	if err != nil {
		return nil, fmt.Errorf("node not found: %w", err)
	}
	_, err = job.ValidateSpec(in.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to validate job spec: %w", err)
	}
	var extractor ExternalJobIDExtractor
	err = toml.Unmarshal([]byte(in.Spec), &extractor)
	if err != nil {
		return nil, fmt.Errorf("failed to load job spec: %w", err)
	}
	if extractor.ExternalJobID == "" {
		return nil, errors.New("externalJobID is required")
	}

	// must auto increment the version to avoid collision on the node side
	proposals, err := j.proposalStore.list(&jobv1.ListProposalsRequest_Filter{
		JobIds: []string{extractor.ExternalJobID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list proposals: %w", err)
	}
	proposalVersion := int32(len(proposals) + 1) //nolint:gosec // G115
	appProposalID, err := n.App.GetFeedsService().ProposeJob(ctx, &feeds.ProposeJobArgs{
		FeedsManagerID: 1,
		Spec:           in.Spec,
		RemoteUUID:     uuid.MustParse(extractor.ExternalJobID),
		Version:        proposalVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to propose job: %w", err)
	}
	fmt.Printf("proposed job uuid %s with id, spec, version: %d\n%s\n%d\n", extractor.ExternalJobID, appProposalID, in.Spec, len(proposals)+1)
	// auto approve for now
	proposedSpec, err := n.App.GetFeedsService().ListSpecsByJobProposalIDs(ctx, []int64{appProposalID})
	if err != nil {
		return nil, fmt.Errorf("failed to list specs: %w", err)
	}
	// possible to have multiple specs for the same job proposal id; take the last one
	if len(proposedSpec) == 0 {
		return nil, fmt.Errorf("no specs found for job proposal id: %d", appProposalID)
	}
	err = n.App.GetFeedsService().ApproveSpec(ctx, proposedSpec[len(proposedSpec)-1].ID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to approve job: %w", err)
	}

	storeProposalID := uuid.Must(uuid.NewRandom()).String()
	p := &jobv1.ProposeJobResponse{Proposal: &jobv1.Proposal{
		// make the proposal id the same as the job id for further reference
		// if you are changing this make sure to change the GetProposal and ListJobs method implementation
		Id:       storeProposalID,
		Revision: int64(proposalVersion),
		// Auto approve for now
		Status:             jobv1.ProposalStatus_PROPOSAL_STATUS_APPROVED,
		DeliveryStatus:     jobv1.ProposalDeliveryStatus_PROPOSAL_DELIVERY_STATUS_DELIVERED,
		Spec:               in.Spec,
		JobId:              extractor.ExternalJobID,
		CreatedAt:          nil,
		UpdatedAt:          nil,
		AckedAt:            nil,
		ResponseReceivedAt: nil,
	}}

	// save the proposal and job
	{
		var (
			storeErr error // used to cleanup if we fail to save the job
			job      *jobv1.Job
		)

		storeErr = j.proposalStore.put(storeProposalID, p.Proposal)
		if err != nil {
			return nil, fmt.Errorf("failed to save proposal: %w", err)
		}
		defer func() {
			// cleanup if we fail to save the job
			if storeErr != nil {
				j.proposalStore.delete(storeProposalID) //nolint:errcheck // ignore error nothing to do
			}
		}()

		job, storeErr = j.jobStore.get(extractor.ExternalJobID)
		if storeErr != nil && !errors.Is(storeErr, errNoExist) {
			return nil, fmt.Errorf("failed to get job: %w", storeErr)
		}
		if errors.Is(storeErr, errNoExist) {
			job = &jobv1.Job{
				Id:          extractor.ExternalJobID,
				Uuid:        extractor.ExternalJobID,
				NodeId:      in.NodeId,
				ProposalIds: []string{storeProposalID},
				Labels:      in.Labels,
			}
		} else {
			job.ProposalIds = append(job.ProposalIds, storeProposalID)
		}
		storeErr = j.jobStore.put(extractor.ExternalJobID, job)
		if storeErr != nil {
			return nil, fmt.Errorf("failed to save job: %w", storeErr)
		}
	}
	return p, nil
}

func (j *JobServiceClient) RevokeJob(ctx context.Context, in *jobv1.RevokeJobRequest, opts ...grpc.CallOption) (*jobv1.RevokeJobResponse, error) {
	// TODO CCIP-3108 implement me
	panic("implement me")
}

func (j *JobServiceClient) DeleteJob(ctx context.Context, in *jobv1.DeleteJobRequest, opts ...grpc.CallOption) (*jobv1.DeleteJobResponse, error) {
	// TODO CCIP-3108 implement me
	panic("implement me")
}

type ExternalJobIDExtractor struct {
	ExternalJobID string `toml:"externalJobID"`
}

var errNoExist = errors.New("does not exist")

// proposalStore is an interface for storing job proposals.
type proposalStore interface {
	put(proposalID string, proposal *jobv1.Proposal) error
	get(proposalID string) (*jobv1.Proposal, error)
	list(filter *jobv1.ListProposalsRequest_Filter) ([]*jobv1.Proposal, error)
	delete(proposalID string) error
}

// jobStore is an interface for storing jobs.
type jobStore interface {
	put(jobID string, job *jobv1.Job) error
	get(jobID string) (*jobv1.Job, error)
	list(filter *jobv1.ListJobsRequest_Filter) ([]*jobv1.Job, error)
	delete(jobID string) error
}

// nodeStore is an interface for storing nodes.
type nodeStore interface {
	put(nodeID string, node *Node) error
	get(nodeID string) (*Node, error)
	list() []*Node
	asMap() map[string]*Node
	delete(nodeID string) error
}

var _ jobStore = &mapJobStore{}

type mapJobStore struct {
	mu            sync.Mutex
	jobs          map[string]*jobv1.Job
	nodesToJobIDs map[string][]string
	uuidToJobIDs  map[string][]string
}

func newMapJobStore() *mapJobStore {
	return &mapJobStore{
		jobs:          make(map[string]*jobv1.Job),
		nodesToJobIDs: make(map[string][]string),
		uuidToJobIDs:  make(map[string][]string),
	}
}

func (m *mapJobStore) put(jobID string, job *jobv1.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.jobs == nil {
		m.jobs = make(map[string]*jobv1.Job)
		m.nodesToJobIDs = make(map[string][]string)
		m.uuidToJobIDs = make(map[string][]string)
	}
	m.jobs[jobID] = job
	if _, ok := m.nodesToJobIDs[job.NodeId]; !ok {
		m.nodesToJobIDs[job.NodeId] = make([]string, 0)
	}
	m.nodesToJobIDs[job.NodeId] = append(m.nodesToJobIDs[job.NodeId], jobID)
	if _, ok := m.uuidToJobIDs[job.Uuid]; !ok {
		m.uuidToJobIDs[job.Uuid] = make([]string, 0)
	}
	m.uuidToJobIDs[job.Uuid] = append(m.uuidToJobIDs[job.Uuid], jobID)
	return nil
}

func (m *mapJobStore) get(jobID string) (*jobv1.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.jobs == nil {
		return nil, fmt.Errorf("%w: job not found: %s", errNoExist, jobID)
	}
	job, ok := m.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("%w: job not found: %s", errNoExist, jobID)
	}
	return job, nil
}

func (m *mapJobStore) list(filter *jobv1.ListJobsRequest_Filter) ([]*jobv1.Job, error) {
	if filter != nil && filter.NodeIds != nil && filter.Uuids != nil && filter.Ids != nil {
		return nil, errors.New("only one of NodeIds, Uuids or Ids can be set")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.jobs == nil {
		return []*jobv1.Job{}, nil
	}

	jobs := make([]*jobv1.Job, 0, len(m.jobs))

	if filter == nil || (filter.NodeIds == nil && filter.Uuids == nil && filter.Ids == nil) {
		for _, job := range m.jobs {
			jobs = append(jobs, job)
		}
		return jobs, nil
	}

	wantedJobIDs := make(map[string]struct{})
	// use node ids to construct wanted job ids
	switch {
	case filter.NodeIds != nil:
		for _, nodeID := range filter.NodeIds {
			jobIDs, ok := m.nodesToJobIDs[nodeID]
			if !ok {
				continue
			}
			for _, jobID := range jobIDs {
				wantedJobIDs[jobID] = struct{}{}
			}
		}
	case filter.Uuids != nil:
		for _, uuid := range filter.Uuids {
			jobIDs, ok := m.uuidToJobIDs[uuid]
			if !ok {
				continue
			}
			for _, jobID := range jobIDs {
				wantedJobIDs[jobID] = struct{}{}
			}
		}
	case filter.Ids != nil:
		for _, jobID := range filter.Ids {
			wantedJobIDs[jobID] = struct{}{}
		}
	default:
		panic("this should never happen because of the nil filter check")
	}

	for _, job := range m.jobs {
		if _, ok := wantedJobIDs[job.Id]; ok {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

func (m *mapJobStore) delete(jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.jobs == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}
	job, ok := m.jobs[jobID]
	if !ok {
		return nil
	}
	delete(m.jobs, jobID)
	delete(m.nodesToJobIDs, job.NodeId)
	delete(m.uuidToJobIDs, job.Uuid)
	return nil
}

var _ proposalStore = &mapProposalStore{}

type mapProposalStore struct {
	mu                sync.Mutex
	proposals         map[string]*jobv1.Proposal
	jobIdToProposalId map[string]string
}

func newMapProposalStore() *mapProposalStore {
	return &mapProposalStore{
		proposals:         make(map[string]*jobv1.Proposal),
		jobIdToProposalId: make(map[string]string),
	}
}

func (m *mapProposalStore) put(proposalID string, proposal *jobv1.Proposal) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.proposals == nil {
		m.proposals = make(map[string]*jobv1.Proposal)
	}
	if m.jobIdToProposalId == nil {
		m.jobIdToProposalId = make(map[string]string)
	}
	m.proposals[proposalID] = proposal
	m.jobIdToProposalId[proposal.JobId] = proposalID
	return nil
}
func (m *mapProposalStore) get(proposalID string) (*jobv1.Proposal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.proposals == nil {
		return nil, fmt.Errorf("proposal not found: %s", proposalID)
	}
	proposal, ok := m.proposals[proposalID]
	if !ok {
		return nil, fmt.Errorf("%w: proposal not found: %s", errNoExist, proposalID)
	}
	return proposal, nil
}
func (m *mapProposalStore) list(filter *jobv1.ListProposalsRequest_Filter) ([]*jobv1.Proposal, error) {
	if filter != nil && filter.GetIds() != nil && filter.GetJobIds() != nil {
		return nil, errors.New("only one of Ids or JobIds can be set")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.proposals == nil {
		return nil, nil
	}
	proposals := make([]*jobv1.Proposal, 0)
	// all proposals
	if filter == nil || (filter.GetIds() == nil && filter.GetJobIds() == nil) {
		for _, proposal := range m.proposals {
			proposals = append(proposals, proposal)
		}
		return proposals, nil
	}

	// can't both be nil at this point
	wantedProposalIDs := filter.GetIds()
	if wantedProposalIDs == nil {
		wantedProposalIDs = make([]string, 0)
		for _, jobId := range filter.GetJobIds() {
			proposalID, ok := m.jobIdToProposalId[jobId]
			if !ok {
				continue
			}
			wantedProposalIDs = append(wantedProposalIDs, proposalID)
		}
	}

	for _, want := range wantedProposalIDs {
		p, ok := m.proposals[want]
		if !ok {
			continue
		}
		proposals = append(proposals, p)
	}
	return proposals, nil
}
func (m *mapProposalStore) delete(proposalID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.proposals == nil {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	delete(m.proposals, proposalID)
	return nil
}

var _ nodeStore = &mapNodeStore{}

type mapNodeStore struct {
	mu    sync.Mutex
	nodes map[string]*Node
}

func newMapNodeStore(n map[string]*Node) *mapNodeStore {
	return &mapNodeStore{
		nodes: n,
	}
}
func (m *mapNodeStore) put(nodeID string, node *Node) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nodes == nil {
		m.nodes = make(map[string]*Node)
	}
	m.nodes[nodeID] = node
	return nil
}
func (m *mapNodeStore) get(nodeID string) (*Node, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nodes == nil {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}
	node, ok := m.nodes[nodeID]
	if !ok {
		return nil, fmt.Errorf("%w: node not found: %s", errNoExist, nodeID)
	}
	return node, nil
}
func (m *mapNodeStore) list() []*Node {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nodes == nil {
		return nil
	}
	nodes := make([]*Node, 0)
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}
func (m *mapNodeStore) delete(nodeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nodes == nil {
		return fmt.Errorf("node not found: %s", nodeID)
	}
	_, ok := m.nodes[nodeID]
	if !ok {
		return nil
	}
	delete(m.nodes, nodeID)
	return nil
}

func (m *mapNodeStore) asMap() map[string]*Node {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nodes == nil {
		return nil
	}
	nodes := make(map[string]*Node)
	for k, v := range m.nodes {
		nodes[k] = v
	}
	return nodes
}
