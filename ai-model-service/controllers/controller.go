package controllers

import (
	"context"
	"fmt"

	"github.com/blcvn/backend/services/ai-model-service/common/constants"
	pb "github.com/blcvn/kratos-proto/go/ai-model"
)

type modelController struct {
	pb.UnimplementedAIModelServiceServer
	usecase   iModelUsecase
	transform iTransform
}

// CreateModel creates a new AI model
func (c *modelController) CreateModel(ctx context.Context, req *pb.CreateModelRequest) (*pb.CreateModelResponse, error) {
	// Transform proto to entity
	payload, err := c.transform.Pb2CreateModelPayload(req.GetPayload())
	if err != nil {
		return &pb.CreateModelResponse{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode_BAD_REQUEST,
				Message: fmt.Sprintf("transform error: %v", err),
			},
		}, nil
	}

	// Call usecase
	model, usecaseErr := c.usecase.CreateModel(ctx, payload)
	if usecaseErr != nil {
		return &pb.CreateModelResponse{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode(usecaseErr.GetCode()),
				Message: usecaseErr.Error(),
			},
		}, nil
	}

	// Transform entity to proto
	modelPb, err := c.transform.Model2Pb(model)
	if err != nil {
		return &pb.CreateModelResponse{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode_INTERNAL,
				Message: fmt.Sprintf("transform error: %v", err),
			},
		}, nil
	}

	return &pb.CreateModelResponse{
		Metadata: req.Metadata,
		Result: &pb.Result{
			Code:    pb.ResultCode_SUCCESS,
			Message: constants.MsgModelCreated,
		},
		Model: modelPb,
	}, nil
}

// GetModel retrieves a model by ID
func (c *modelController) GetModel(ctx context.Context, req *pb.GetModelRequest) (*pb.GetModelResponse, error) {
	model, err := c.usecase.GetModel(ctx, req.GetId())
	if err != nil {
		return &pb.GetModelResponse{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode(err.GetCode()),
				Message: err.Error(),
			},
		}, nil
	}

	modelPb, transformErr := c.transform.Model2Pb(model)
	if transformErr != nil {
		return &pb.GetModelResponse{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode_INTERNAL,
				Message: fmt.Sprintf("transform error: %v", transformErr),
			},
		}, nil
	}

	return &pb.GetModelResponse{
		Metadata: req.Metadata,
		Result: &pb.Result{
			Code:    pb.ResultCode_SUCCESS,
			Message: constants.MsgModelRetrieved,
		},
		Model: modelPb,
	}, nil
}

// ListModels lists models with filtering
func (c *modelController) ListModels(ctx context.Context, req *pb.ListModelsRequest) (*pb.ListModelsResponse, error) {
	filter, err := c.transform.Pb2ModelFilter(
		req.GetFilter().GetProvider(),
		req.GetFilter().GetStatus(),
		req.GetFilter().GetPage(),
		req.GetFilter().GetPageSize(),
	)
	if err != nil {
		return &pb.ListModelsResponse{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode_BAD_REQUEST,
				Message: fmt.Sprintf("transform error: %v", err),
			},
		}, nil
	}

	models, total, usecaseErr := c.usecase.ListModels(ctx, filter)
	if usecaseErr != nil {
		return &pb.ListModelsResponse{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode(usecaseErr.GetCode()),
				Message: usecaseErr.Error(),
			},
		}, nil
	}

	// Transform models
	modelsPb := make([]*pb.AIModel, 0, len(models))
	for _, model := range models {
		modelPb, err := c.transform.Model2Pb(model)
		if err != nil {
			continue
		}
		modelsPb = append(modelsPb, modelPb)
	}

	return &pb.ListModelsResponse{
		Metadata: req.Metadata,
		Result: &pb.Result{
			Code:    pb.ResultCode_SUCCESS,
			Message: constants.MsgModelsListed,
		},
		Models: modelsPb,
		Total:  int32(total),
	}, nil
}

// UpdateModel updates an existing model
func (c *modelController) UpdateModel(ctx context.Context, req *pb.UpdateModelRequest) (*pb.UpdateModelResponse, error) {
	payload, err := c.transform.Pb2UpdateModelPayload(req.GetPayload())
	if err != nil {
		return &pb.UpdateModelResponse{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode_BAD_REQUEST,
				Message: fmt.Sprintf("transform error: %v", err),
			},
		}, nil
	}

	model, usecaseErr := c.usecase.UpdateModel(ctx, payload)
	if usecaseErr != nil {
		return &pb.UpdateModelResponse{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode(usecaseErr.GetCode()),
				Message: usecaseErr.Error(),
			},
		}, nil
	}

	modelPb, err := c.transform.Model2Pb(model)
	if err != nil {
		return &pb.UpdateModelResponse{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode_INTERNAL,
				Message: fmt.Sprintf("transform error: %v", err),
			},
		}, nil
	}

	return &pb.UpdateModelResponse{
		Metadata: req.Metadata,
		Result: &pb.Result{
			Code:    pb.ResultCode_SUCCESS,
			Message: constants.MsgModelUpdated,
		},
		Model: modelPb,
	}, nil
}

// DeleteModel deletes a model
func (c *modelController) DeleteModel(ctx context.Context, req *pb.DeleteModelRequest) (*pb.ResponseEmpty, error) {
	err := c.usecase.DeleteModel(ctx, req.GetId())
	if err != nil {
		// Note: We can't return error details in Empty response
		// Consider logging here
		return &pb.ResponseEmpty{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode(err.GetCode()),
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.ResponseEmpty{
		Metadata: req.Metadata,
		Result: &pb.Result{
			Code:    pb.ResultCode_SUCCESS,
			Message: constants.MsgModelDeleted,
		},
	}, nil
}

// GetCredentials retrieves API credentials from Vault (internal gRPC only)
func (c *modelController) GetCredentials(ctx context.Context, req *pb.GetCredentialsRequest) (*pb.GetCredentialsResponse, error) {
	creds, err := c.usecase.GetCredentials(ctx, req.GetModelId())
	if err != nil {
		return &pb.GetCredentialsResponse{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode(err.GetCode()),
				Message: err.Error(),
			},
		}, nil
	}

	credsPb, transformErr := c.transform.Credentials2Pb(creds)
	if transformErr != nil {
		return &pb.GetCredentialsResponse{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode_INTERNAL,
				Message: fmt.Sprintf("transform error: %v", transformErr),
			},
		}, nil
	}

	return &pb.GetCredentialsResponse{
		Metadata: req.Metadata,
		Result: &pb.Result{
			Code:    pb.ResultCode_SUCCESS,
			Message: constants.MsgCredentialsRetrieved,
		},
		Credentials: credsPb,
	}, nil
}

// LogUsage logs AI usage (internal gRPC only)
func (c *modelController) LogUsage(ctx context.Context, req *pb.LogUsageRequest) (*pb.ResponseEmpty, error) {
	payload, err := c.transform.Pb2LogUsagePayload(req.GetPayload())
	if err != nil {
		return &pb.ResponseEmpty{
			Result: &pb.Result{
				Code:    pb.ResultCode_BAD_REQUEST,
				Message: fmt.Sprintf("transform error: %v", err),
			},
		}, nil
	}

	usecaseErr := c.usecase.LogUsage(ctx, payload)
	if usecaseErr != nil {
		return &pb.ResponseEmpty{
			Result: &pb.Result{
				Code:    pb.ResultCode_INTERNAL, // Map error code appropriately
				Message: usecaseErr.Error(),
			},
		}, nil
	}

	return &pb.ResponseEmpty{
		Result: &pb.Result{
			Code:    pb.ResultCode_SUCCESS,
			Message: "Usage logged",
		},
	}, nil
}

// CheckQuota checks quota limits (internal gRPC only)
func (c *modelController) CheckQuota(ctx context.Context, req *pb.CheckQuotaRequest) (*pb.CheckQuotaResponse, error) {
	quota, err := c.usecase.CheckQuota(ctx, req.GetModelId())
	if err != nil {
		return &pb.CheckQuotaResponse{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode(err.GetCode()),
				Message: err.Error(),
			},
		}, nil
	}

	quotaPb, transformErr := c.transform.QuotaStatus2Pb(quota)
	if transformErr != nil {
		return &pb.CheckQuotaResponse{
			Metadata: req.Metadata,
			Result: &pb.Result{
				Code:    pb.ResultCode_INTERNAL,
				Message: fmt.Sprintf("transform error: %v", transformErr),
			},
		}, nil
	}

	return &pb.CheckQuotaResponse{
		Metadata: req.Metadata,
		Result: &pb.Result{
			Code:    pb.ResultCode_SUCCESS,
			Message: constants.MsgQuotaChecked,
		},
		Quota: quotaPb,
	}, nil
}

// GetUsageStats retrieves usage statistics
func (c *modelController) GetUsageStats(ctx context.Context, req *pb.GetUsageStatsRequest) (*pb.GetUsageStatsResponse, error) {
	// TODO: Implement usage stats aggregation
	return &pb.GetUsageStatsResponse{
		Metadata: req.Metadata,
		Result: &pb.Result{
			Code:    pb.ResultCode_SUCCESS,
			Message: "Usage stats retrieved",
		},
		Stats: []*pb.UsageStats{},
	}, nil
}
