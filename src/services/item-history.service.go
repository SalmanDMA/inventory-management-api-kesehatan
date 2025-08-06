package services

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ItemHistoryService struct {
	ItemHistoryRepository repositories.ItemHistoryRepository
	ItemRepository        repositories.ItemRepository
}

func NewItemHistoryService(itemHistoryRepo repositories.ItemHistoryRepository, itemRepo repositories.ItemRepository) *ItemHistoryService {
	return &ItemHistoryService{
		ItemHistoryRepository: itemHistoryRepo,
		ItemRepository:        itemRepo,
	}
}

func (service *ItemHistoryService) GetAllItemHistories(itemID uuid.UUID, changeType string) ([]models.ResponseGetItemHistory, error) {
	itemHistories, err := service.ItemHistoryRepository.FindAllByItem(itemID, changeType)
	if err != nil {
		return nil, err
	}

	var itemHistoriesResponse []models.ResponseGetItemHistory
	for _, itemHistory := range itemHistories {
		itemHistoriesResponse = append(itemHistoriesResponse, models.ResponseGetItemHistory{
			ID:          itemHistory.ID,
			ItemID:      itemHistory.ItemID,
			ChangeType:  itemHistory.ChangeType,
			Description: itemHistory.Description,
			OldPrice:    itemHistory.OldPrice,
			NewPrice:    itemHistory.NewPrice,
			CurrentPrice: itemHistory.CurrentPrice,
			OldStock:    itemHistory.OldStock,
			NewStock:    itemHistory.NewStock,
			CurrentStock: itemHistory.CurrentStock,
			CreatedBy:   *itemHistory.CreatedBy,
			UpdatedBy:   *itemHistory.UpdatedBy,
			DeletedBy:   *itemHistory.DeletedBy,
			Item:        itemHistory.Item,
			CreatedByUser: itemHistory.CreatedByUser,
			UpdatedByUser: itemHistory.UpdatedByUser,
			CreatedAt: 	itemHistory.CreatedAt,
			UpdatedAt: 	itemHistory.UpdatedAt,
			DeletedAt: 	itemHistory.DeletedAt,
		})
	}

	return itemHistoriesResponse, nil
}

func (service *ItemHistoryService) GetAllItemHistoriesPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.ItemHistoryPaginatedResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100 
	}
	if req.Status == "" {
		req.Status = "active"
	}

	itemHistories, totalCount, err := service.ItemHistoryRepository.FindAllPaginated(req)
	if err != nil {
		return nil, err
	}

	itemHistoriesResponse := []models.ResponseGetItemHistory{}
	for _, itemHistory := range itemHistories {
		itemHistoriesResponse = append(itemHistoriesResponse, models.ResponseGetItemHistory{
			ID:            itemHistory.ID,
			ItemID:        itemHistory.ItemID,
			ChangeType:    itemHistory.ChangeType,
			Description:   itemHistory.Description,
			OldPrice:      itemHistory.OldPrice,
			NewPrice:      itemHistory.NewPrice,
			CurrentPrice:  itemHistory.CurrentPrice,
			OldStock:      itemHistory.OldStock,
			NewStock:      itemHistory.NewStock,
			CurrentStock:  itemHistory.CurrentStock,
			CreatedBy: func() uuid.UUID {
				if itemHistory.CreatedBy != nil {
					return *itemHistory.CreatedBy
				}
				return uuid.Nil
			}(),
			UpdatedBy: func() uuid.UUID {
				if itemHistory.UpdatedBy != nil {
					return *itemHistory.UpdatedBy
				}
				return uuid.Nil
			}(),
			DeletedBy: func() uuid.UUID {
				if itemHistory.DeletedBy != nil {
					return *itemHistory.DeletedBy
				}
				return uuid.Nil
			}(),
			Item:           itemHistory.Item,
			CreatedByUser:  itemHistory.CreatedByUser,
			UpdatedByUser:  itemHistory.UpdatedByUser,
			CreatedAt:      itemHistory.CreatedAt,
			UpdatedAt:      itemHistory.UpdatedAt,
			DeletedAt:      itemHistory.DeletedAt,
		})
	}

	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	paginationResponse := models.PaginationResponse{
		CurrentPage:  req.Page,
		PerPage:      req.Limit,
		TotalPages:   totalPages,
		TotalRecords: totalCount,
		HasNext:      hasNext,
		HasPrev:      hasPrev,
	}

	return &models.ItemHistoryPaginatedResponse{
		Data:       itemHistoriesResponse,
		Pagination: paginationResponse,
	}, nil
}

func (service *ItemHistoryService) CreateItemHistory(itemHistoryRequest *models.ItemHistoryCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.ItemHistory, error) {
	changeType := strings.ToLower(itemHistoryRequest.ChangeType)
	lastHistory, _ := service.ItemHistoryRepository.FindLastByItem(itemHistoryRequest.ItemID, changeType)

	var oldPrice, oldStock int
	if lastHistory != nil {
		oldPrice = lastHistory.CurrentPrice
		oldStock = lastHistory.CurrentStock
	}

	newItemHistory := &models.ItemHistory{
		ID:           uuid.New(),
		ItemID:       itemHistoryRequest.ItemID,
		ChangeType:   changeType,
		Description:  itemHistoryRequest.Description,
		CreatedBy:    &userInfo.ID,
		UpdatedBy:    &userInfo.ID,
		OldPrice:     oldPrice,
		OldStock:     oldStock,
		CurrentPrice: oldPrice,
		CurrentStock: oldStock,
	}

	switch changeType {
	case "create_price", "update_price":
		newItemHistory.NewPrice = itemHistoryRequest.NewPrice
		newItemHistory.CurrentPrice = itemHistoryRequest.NewPrice
	case "create_stock", "update_stock":
		newItemHistory.NewStock = itemHistoryRequest.NewStock
		newItemHistory.CurrentStock = itemHistoryRequest.NewStock
	default:
		return nil, fmt.Errorf("invalid change_type '%s'", changeType)
	}

	itemHistory, err := service.ItemHistoryRepository.Insert(newItemHistory)
	if err != nil {
		return nil, err
	}

	item, err := service.ItemRepository.FindById(itemHistoryRequest.ItemID.String(), false)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("item not found")
	}

	switch changeType {
	case "update_price", "create_price":
		item.Price = itemHistoryRequest.NewPrice
	case "update_stock", "create_stock":
		item.Stock = itemHistoryRequest.NewStock
	}

	if _, err := service.ItemRepository.Update(item); err != nil {
		return nil, err
	}

	return itemHistory, nil
}

func (service *ItemHistoryService) DeleteItemHistories(itemHistoryRequest *models.ItemHistoryIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, itemHistoryId := range itemHistoryRequest.IDs {
		_, err := service.ItemHistoryRepository.FindById(itemHistoryId.String(), false)
		if err != nil {
			if err == repositories.ErrItemHistoryNotFound {
				log.Printf("ItemHistory not found: %v\n", itemHistoryId)
				continue
			}
			log.Printf("Error finding itemHistory %v: %v\n", itemHistoryId, err)
			return errors.New("error finding useitemHistoryr")
		}

		if itemHistoryRequest.IsHardDelete == "hardDelete" {
			if err := service.ItemHistoryRepository.Delete(itemHistoryId.String(), true); err != nil {
				log.Printf("Error hard deleting itemHistory %v: %v\n", itemHistoryId, err)
				return errors.New("error hard deleting itemHistory")
			}
		} else {
			if err := service.ItemHistoryRepository.Delete(itemHistoryId.String(), false); err != nil {
				log.Printf("Error soft deleting itemHistory %v: %v\n", itemHistoryId, err)
				return errors.New("error soft deleting itemHistory")
			}
		}
	}

	return nil
}

func (service *ItemHistoryService) RestoreItemHistories(itemHistory *models.ItemHistoryRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.ItemHistory, error) {
	var restoredItemHistories []models.ItemHistory

	for _, itemHistoryId := range itemHistory.IDs {
		itemHistory := &models.ItemHistory{ID: itemHistoryId}

		restoredItemHistory, err := service.ItemHistoryRepository.Restore(itemHistory, itemHistoryId.String())
		if err != nil {
			if err == repositories.ErrItemHistoryNotFound {
				log.Printf("ItemHistory not found: %v\n", itemHistoryId)
				continue
			}
			log.Printf("Error restoring itemHistory %v: %v\n", itemHistoryId, err)
			return nil, errors.New("error restoring itemHistory")
		}

		restoredItemHistories = append(restoredItemHistories, *restoredItemHistory)
	}

	return restoredItemHistories, nil
}