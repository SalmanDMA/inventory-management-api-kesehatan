package helpers

// import (
// 	"fmt"

// 	"github.com/SalmanDMA/inventory-app/backend/src/models"
// 	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
// 	"github.com/google/uuid"
// )

// type ItemStockHelper struct {
// 	ItemRepo        repositories.ItemRepository
// 	ItemStockRepo   repositories.ItemStockRepository
// 	ItemHistoryRepo repositories.ItemHistoryRepository
// }

// func NewItemStockHelper(itemRepo repositories.ItemRepository, itemStockRepo repositories.ItemStockRepository, itemHistoryRepo repositories.ItemHistoryRepository) *ItemStockHelper {
// 	return &ItemStockHelper{ItemRepo: itemRepo, ItemStockRepo: itemStockRepo, ItemHistoryRepo: itemHistoryRepo}
// }

// func (h *ItemStockHelper) CreateOrUpdateItemStock(
// 	itemID uuid.UUID,
// 	warehouseID uuid.UUID,
// 	newStock int,
// 	lowStock int,
// 	changeType string,
// 	description string,
// ) error {
// 	lastStock, err := h.ItemStockRepo.FindByItemAndWarehouseId(itemID, warehouseID)
// 	if err != nil && err != repositories.ErrItemStockNotFound {
// 		return fmt.Errorf("failed to fetch item stock: %w", err)
// 	}

// 	lastHistory, err := h.ItemHistoryRepo.FindByItemAndWarehouseId(itemID, warehouseID, changeType)
// 	if err != nil && err != repositories.ErrItemHistoryNotFound {
// 		return fmt.Errorf("failed to fetch item history: %w", err)
// 	}

// 	var oldStock, stockDelta int
// 	if lastHistory != nil {
// 		oldStock = lastHistory.CurrentStock
// 	}

// 	if changeType == "CREATE_STOCK" {
// 		oldStock = 0
// 	}

// 	if changeType == "STOCK_OUT" || changeType == "STOCK_OUT_UPDATE" {
// 	 stockDelta = oldStock - newStock
// 	} else {
// 		stockDelta = oldStock + newStock
// 	}

// 	if lastStock == nil {
// 		newStockModel := &models.ItemStock{
// 			ID:          uuid.New(),
// 			ItemID:      itemID,
// 			WarehouseID: warehouseID,
// 			Stock:       newStock,
// 			LowStock:    lowStock,
// 		}
// 		if _, err := h.ItemStockRepo.Insert(newStockModel); err != nil {
// 			return fmt.Errorf("failed to insert item stock: %w", err)
// 		}
// 	} else {
// 		lastStock.Stock = stockDelta
// 		lastStock.LowStock = lowStock
// 		fmt.Println(lastStock, "lastStock")
// 		if _, err := h.ItemStockRepo.Update(lastStock); err != nil {
// 			return fmt.Errorf("failed to update item stock: %w", err)
// 		}
// 	}

// 	newHistory := &models.ItemHistory{
// 		ID:          uuid.New(),
// 		ItemID:      itemID,
// 		WarehouseID: warehouseID,
// 		ChangeType:  changeType,
// 		Description: description,
// 		OldStock:    oldStock,
// 		NewStock:    newStock,
// 		CurrentStock:  stockDelta,
// 		OldPrice:    0,
// 		NewPrice:    0,
// 		CurrentPrice:  0,
// 	}

// 	fmt.Println(newHistory, "newHistory")

// 	if _, err := h.ItemHistoryRepo.Insert(newHistory); err != nil {
// 		return fmt.Errorf("failed to insert item history: %w", err)
// 	}

// 	return nil
// }

// func (h *ItemStockHelper) UpdateItemPrice(
// 	itemID uuid.UUID,
// 	warehouseID uuid.UUID,
// 	newPrice int,
// 	changeType string,
// 	description string,
// ) error {
// 	lastHistory, err := h.ItemHistoryRepo.FindByItemAndWarehouseId(itemID, warehouseID, changeType)
// 	if err != nil && err != repositories.ErrItemHistoryNotFound {
// 		return fmt.Errorf("failed to fetch item history: %w", err)
// 	}

// 	var oldPrice int
// 	if lastHistory != nil {
// 		oldPrice = lastHistory.CurrentPrice
// 	}

// 	if changeType == "CREATE_PRICE" {
// 		oldPrice = 0
// 	}

// 	 currentPrice := oldPrice + newPrice

// 		newHistory := &models.ItemHistory{
// 		ID:          uuid.New(),
// 		ItemID:      itemID,
// 		WarehouseID: warehouseID,
// 		ChangeType:  changeType,
// 		Description: description,
// 		OldStock:    0,
// 		NewStock:    0,
// 		CurrentStock:  0,
// 		OldPrice:    oldPrice,
// 		NewPrice:    newPrice,
// 		CurrentPrice:  currentPrice,
// 	}

// 	if _, err := h.ItemHistoryRepo.Insert(newHistory); err != nil {
// 		return fmt.Errorf("failed to insert item history: %w", err)
// 	}

// 	return nil
// }

// func (h *ItemStockHelper) StockIn(
// 	itemID uuid.UUID,
// 	warehouseID uuid.UUID,
// 	quantity int,
// 	lowStock int,
// 	description string,
// 	changeType string,
// ) error {
// 	_, err := h.ItemStockRepo.FindByItemAndWarehouseId(itemID, warehouseID)
// 	if err != nil && err != repositories.ErrItemStockNotFound {
// 		return fmt.Errorf("failed to fetch current stock: %w", err)
// 	}

// 	return h.CreateOrUpdateItemStock(itemID, warehouseID, quantity, lowStock, changeType, description)
// }

// func (h *ItemStockHelper) StockOut(
// 	itemID uuid.UUID,
// 	warehouseID uuid.UUID,
// 	quantity int,
// 	lowStock int,
// 	description string,
// 	changeType string,
// ) error {
// 	_, err := h.ItemStockRepo.FindByItemAndWarehouseId(itemID, warehouseID)
// 	if err != nil {
// 		return fmt.Errorf("failed to fetch current stock: %w", err)
// 	}

// 	return h.CreateOrUpdateItemStock(itemID, warehouseID, quantity, lowStock, changeType, description)
// }