package usecases

type InventoryUsecase interface {
}

type InventoryUseCase struct {
}

func NewInventoryUsecase() InventoryUsecase {
	return &InventoryUseCase{}
}