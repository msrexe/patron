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

	wrk *worker
}

func TestWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerTestSuite))
}

func (suite *WorkerTestSuite) SetupTest() {
	workers := newWorkerArray(func(job *Job) error {
		time.Sleep(1 * time.Second)
		payloadName, err := job.GetPayload("name")
		if err != nil {
			return err
		}

		fmt.Printf("%d. job completed.\nJob payload name: %s\n", job.ID, payloadName)

		return nil
	}, 1)

	suite.wrk = workers[0]
}

func (suite *WorkerTestSuite) TestNewWorkerArray() {
	suite.Len(newWorkerArray(nil, 2), 2)
}

func (suite *WorkerTestSuite) TestGetID() {
	suite.Equal(suite.wrk.GetID(), 4)
}

func (suite *WorkerTestSuite) TestGetJob() {
	suite.wrk.SetJob(&Job{
		ID: 10,
	})

	suite.Equal(suite.wrk.GetJob().ID, 10)
}

func (suite *WorkerTestSuite) TestSetJob() {
	suite.wrk.SetJob(&Job{
		ID:      10,
		Context: context.Background(),
		Payload: map[string]interface{}{
			"name":     "HTTP Request",
			"dest_url": "http://localhost:8080/",
		},
	})

	suite.NotEmpty(suite.wrk.GetJob())
}

func (suite *WorkerTestSuite) TestFinalizeJob() {
	suite.wrk.SetJob(&Job{
		ID: 10,
	})
	suite.NotEmpty(suite.wrk.GetJob())

	suite.wrk.FinalizeJob()
	suite.Empty(suite.wrk.GetJob())
}

func (suite *WorkerTestSuite) TestIsBusy() {
	suite.False(suite.wrk.IsBusy())

	suite.wrk.SetJob(&Job{
		ID: 10,
	})
	suite.True(suite.wrk.IsBusy())
}

func (suite *WorkerTestSuite) TestWorkSuccess() {
	suite.wrk.SetJob(&Job{
		ID:      10,
		Context: context.Background(),
		Payload: map[string]interface{}{
			"name":     "HTTP Request",
			"dest_url": "http://localhost:8080/",
		},
	})

	suite.NoError(suite.wrk.Work())
}

func (suite *WorkerTestSuite) TestWorkFailure() {
	suite.wrk.SetJob(&Job{
		ID:      10,
		Context: context.Background(),
		Payload: map[string]interface{}{},
	})

	suite.ErrorIs(suite.wrk.Work(), ErrJobPayloadNotFound)
}
