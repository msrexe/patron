package patron

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type WorkerTestSuite struct {
	suite.Suite

	worker Worker
}

func TestWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerTestSuite))
}

func (suite *WorkerTestSuite) SetupTest() {
	workers := NewWorkerArray(func(job *Job) error {
		time.Sleep(1 * time.Second)
		payloadName, err := job.GetPayload("name")
		if err != nil {
			return err
		}

		fmt.Printf("%d. job completed.\nJob payload name: %s\n", job.ID, payloadName)

		return nil
	}, 5)

	suite.worker = workers[4]
}

func (suite *WorkerTestSuite) TestNewWorkerArray() {
	suite.Len(NewWorkerArray(nil, 2), 2)
}

func (suite *WorkerTestSuite) TestGetID() {
	suite.Equal(suite.worker.GetID(), 4)
}

func (suite *WorkerTestSuite) TestGetJob() {
	suite.worker.SetJob(&Job{
		ID: 10,
	})

	suite.Equal(suite.worker.GetJob().ID, 10)
}

func (suite *WorkerTestSuite) TestSetJob() {
	suite.worker.SetJob(&Job{
		ID:      10,
		Context: context.Background(),
		Payload: map[string]interface{}{
			"name":     "HTTP Request",
			"dest_url": "http://localhost:8080/",
		},
	})

	suite.NotEmpty(suite.worker.GetJob())
}

func (suite *WorkerTestSuite) TestFinalizeJob() {
	suite.worker.SetJob(&Job{
		ID: 10,
	})
	suite.NotEmpty(suite.worker.GetJob())

	suite.worker.FinalizeJob()
	suite.Empty(suite.worker.GetJob())
}

func (suite *WorkerTestSuite) TestIsBusy() {
	suite.False(suite.worker.IsBusy())

	suite.worker.SetJob(&Job{
		ID: 10,
	})
	suite.True(suite.worker.IsBusy())
}

func (suite *WorkerTestSuite) TestWorkSuccess() {
	suite.worker.SetJob(&Job{
		ID:      10,
		Context: context.Background(),
		Payload: map[string]interface{}{
			"name":     "HTTP Request",
			"dest_url": "http://localhost:8080/",
		},
	})

	suite.NoError(suite.worker.Work())
}

func (suite *WorkerTestSuite) TestWorkFailure() {
	suite.worker.SetJob(&Job{
		ID:      10,
		Context: context.Background(),
		Payload: map[string]interface{}{},
	})

	suite.ErrorIs(suite.worker.Work(), ErrJobPayloadNotFound)
}
