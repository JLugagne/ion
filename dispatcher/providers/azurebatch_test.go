package providers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/Azure/azure-sdk-for-go/services/batch/2017-09-01.6.0/batch"
	"github.com/lawrencegripper/ion/dispatcher/messaging"
	"github.com/lawrencegripper/ion/dispatcher/types"
	apiv1 "k8s.io/api/core/v1"
)

const (
	mockDispatcherName = "mockdispatchername"
	mockMessageID      = "examplemessageID"
)

func NewMockAzureBatchProvider(createTask func(taskDetails batch.TaskAddParameter) (autorest.Response, error), listTasks func() (*[]batch.CloudTask, error)) (*AzureBatch, error) {
	b := AzureBatch{}

	b.jobConfig = &types.JobConfig{
		SidecarImage: "sidecar-image",
		WorkerImage:  "worker-image",
	}
	b.dispatcherName = mockDispatcherName

	b.inprogressJobStore = map[string]messaging.Message{}
	b.createTask = createTask
	b.listTasks = listTasks
	return &b, nil
}

func TestPod2DockerGeneratesValidOutputEncoding(t *testing.T) {
	containers := []apiv1.Container{
		{
			Name:  "sidecar",
			Image: "barry",
			Args:  []string{"encoding"},
		},
		{
			Name:            "worker",
			Image:           "marge",
			ImagePullPolicy: apiv1.PullAlways,
		},
	}

	// Todo: Pull this out into a standalone package once stabilized
	podCommand, err := getPodCommand(batchPodComponents{
		Containers: containers,
		PodName:    mockMessageID,
		TaskID:     mockMessageID,
		Volumes:    nil,
	})

	if err != nil {
		t.Error(err)
	}

	t.Log(podCommand)
	if strings.Contains(podCommand, "&lt;") {
		t.Error("output contains incorrect encoding")
	}
}

func TestAzureBatchDispatchAddsJob(t *testing.T) {
	inMemMockTaskStore := []batch.CloudTask{}

	create := func(taskDetails batch.TaskAddParameter) (autorest.Response, error) {
		inMemMockTaskStore = append(inMemMockTaskStore, batch.CloudTask{})
		return autorest.Response{}, nil
	}

	list := func() (*[]batch.CloudTask, error) {
		return &inMemMockTaskStore, nil
	}

	b, _ := NewMockAzureBatchProvider(create, list)

	messageToSend := MockMessage{
		MessageID: mockMessageID,
	}

	err := b.Dispatch(messageToSend)

	if err != nil {
		t.Error(err)
	}

	jobsLen := len(inMemMockTaskStore)
	if jobsLen != 1 {
		t.Errorf("Job count incorrected Expected: 1 Got: %v", jobsLen)
	}
}

func TestAzureBatchFailedDispatchRejectsMessage(t *testing.T) {
	inMemMockTaskStore := []batch.CloudTask{}

	create := func(taskDetails batch.TaskAddParameter) (autorest.Response, error) {
		return autorest.Response{}, fmt.Errorf("Simulate error")
	}

	list := func() (*[]batch.CloudTask, error) {
		return &inMemMockTaskStore, nil
	}

	b, _ := NewMockAzureBatchProvider(create, list)

	wasRejected := false
	messageToSend := MockMessage{
		MessageID: mockMessageID,
	}
	messageToSend.Rejected = func() {
		wasRejected = true
	}

	err := b.Dispatch(messageToSend)

	if err == nil {
		t.Error("Expected error ... didn't see one!")
	}

	if !wasRejected {
		t.Error("Expected to be rejected... wasn't")
	}

	if len(inMemMockTaskStore) > 0 {
		t.Error("Expected job to not be stored")
	}
}

func TestAzureBatchReconcileJobCompleted(t *testing.T) {
	//Setup... it's a long one. We need to schedule a job first
	inMemMockTaskStore := []batch.CloudTask{}

	create := func(taskDetails batch.TaskAddParameter) (autorest.Response, error) {
		inMemMockTaskStore = append(inMemMockTaskStore, batch.CloudTask{
			ID:          to.StringPtr(mockMessageID),
			DisplayName: to.StringPtr(mockDispatcherName),
		})
		return autorest.Response{}, nil
	}

	list := func() (*[]batch.CloudTask, error) {
		return &inMemMockTaskStore, nil
	}

	b, _ := NewMockAzureBatchProvider(create, list)

	wasAccepted := false
	messageToSend := MockMessage{
		MessageID: mockMessageID,
	}
	messageToSend.Accepted = func() {
		wasAccepted = true
	}
	messageToSend.Rejected = func() {
		t.Error("Message rejected unexpectedly")
	}

	err := b.Dispatch(messageToSend)
	if err != nil {
		t.Error(err)
	}

	task := &inMemMockTaskStore[0]
	task.State = "completed"
	task.ExecutionInfo = &batch.TaskExecutionInformation{
		ExitCode: to.Int32Ptr(0),
	}
	//Lets test things...
	err = b.Reconcile()
	if err != nil {
		t.Error(err)
	}

	if !wasAccepted {
		t.Error("Failed to accept message during reconcilation. Expected message to be marked as accepted as job is complete")
	}

	if b.InProgressCount() != 0 {
		t.Error("Reconcile should remove jobs from the inmemory store once it has accepted or rejected them")
	}
}

func TestAzureBatchReconcileJobFailed(t *testing.T) {
	//Setup... it's a long one. We need to schedule a job first
	inMemMockTaskStore := []batch.CloudTask{}

	create := func(taskDetails batch.TaskAddParameter) (autorest.Response, error) {
		inMemMockTaskStore = append(inMemMockTaskStore, batch.CloudTask{
			ID: to.StringPtr(mockMessageID),
		})
		return autorest.Response{}, nil
	}

	list := func() (*[]batch.CloudTask, error) {
		return &inMemMockTaskStore, nil
	}

	b, _ := NewMockAzureBatchProvider(create, list)

	var rejectedMessage bool

	messageToSend := MockMessage{
		MessageID: mockMessageID,
		Rejected: func() {
			rejectedMessage = true
		},
	}

	err := b.Dispatch(messageToSend)
	if err != nil {
		t.Error(err)
	}

	task := &inMemMockTaskStore[0]
	task.State = "completed"
	task.ExecutionInfo = &batch.TaskExecutionInformation{
		ExitCode: to.Int32Ptr(127),
	}
	//Lets test things...
	err = b.Reconcile()
	if err != nil {
		t.Error(err)
	}

	if !rejectedMessage {
		t.Error("Failed to accept message during reconcilation. Expected message to be marked as accepted as job is complete")
	}

	if b.InProgressCount() != 0 {
		t.Error("Reconcile should remove jobs from the inmemory store once it has accepted or rejected them")
	}
}