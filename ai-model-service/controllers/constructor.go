package controllers

// NewModelController creates a new model controller
func NewModelController(
	usecase iModelUsecase,
	transform iTransform,
) *modelController {
	return &modelController{
		usecase:   usecase,
		transform: transform,
	}
}
