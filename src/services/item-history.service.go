package services

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

// ==============================
// Reads (tanpa transaksi)
// ==============================

func (service *ItemHistoryService) GetAllItemHistories(itemID uuid.UUID, changeType string) ([]models.ResponseGetItemHistory, error) {
	itemHistories, err := service.ItemHistoryRepository.FindAllByItem(nil, itemID, changeType)
	if err != nil {
		return nil, err
	}

	resp := make([]models.ResponseGetItemHistory, 0, len(itemHistories))
	for _, h := range itemHistories {
		resp = append(resp, models.ResponseGetItemHistory{
			ID:            h.ID,
			ItemID:        h.ItemID,
			ChangeType:    h.ChangeType,
			Description:   h.Description,
			OldPrice:      h.OldPrice,
			NewPrice:      h.NewPrice,
			CurrentPrice:  h.CurrentPrice,
			OldStock:      h.OldStock,
			NewStock:      h.NewStock,
			CurrentStock:  h.CurrentStock,
			CreatedBy: func() uuid.UUID {
				if h.CreatedBy != nil {
					return *h.CreatedBy
				}
				return uuid.Nil
			}(),
			UpdatedBy: func() uuid.UUID {
				if h.UpdatedBy != nil {
					return *h.UpdatedBy
				}
				return uuid.Nil
			}(),
			DeletedBy: func() uuid.UUID {
				if h.DeletedBy != nil {
					return *h.DeletedBy
				}
				return uuid.Nil
			}(),
			Item:           h.Item,
			CreatedByUser:  h.CreatedByUser,
			UpdatedByUser:  h.UpdatedByUser,
			CreatedAt:      h.CreatedAt,
			UpdatedAt:      h.UpdatedAt,
			DeletedAt:      h.DeletedAt,
		})
	}
	return resp, nil
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

	list, totalCount, err := service.ItemHistoryRepository.FindAllPaginated(nil, req)
	if err != nil {
		return nil, err
	}

	data := make([]models.ResponseGetItemHistory, 0, len(list))
	for _, h := range list {
		data = append(data, models.ResponseGetItemHistory{
			ID:            h.ID,
			ItemID:        h.ItemID,
			ChangeType:    h.ChangeType,
			Description:   h.Description,
			OldPrice:      h.OldPrice,
			NewPrice:      h.NewPrice,
			CurrentPrice:  h.CurrentPrice,
			OldStock:      h.OldStock,
			NewStock:      h.NewStock,
			CurrentStock:  h.CurrentStock,
			CreatedBy: func() uuid.UUID {
				if h.CreatedBy != nil {
					return *h.CreatedBy
				}
				return uuid.Nil
			}(),
			UpdatedBy: func() uuid.UUID {
				if h.UpdatedBy != nil {
					return *h.UpdatedBy
				}
				return uuid.Nil
			}(),
			DeletedBy: func() uuid.UUID {
				if h.DeletedBy != nil {
					return *h.DeletedBy
				}
				return uuid.Nil
			}(),
			Item:           h.Item,
			CreatedByUser:  h.CreatedByUser,
			UpdatedByUser:  h.UpdatedByUser,
			CreatedAt:      h.CreatedAt,
			UpdatedAt:      h.UpdatedAt,
			DeletedAt:      h.DeletedAt,
		})
	}

	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	return &models.ItemHistoryPaginatedResponse{
		Data: data,
		Pagination: models.PaginationResponse{
			CurrentPage:  req.Page,
			PerPage:      req.Limit,
			TotalPages:   totalPages,
			TotalRecords: totalCount,
			HasNext:      req.Page < totalPages,
			HasPrev:      req.Page > 1,
		},
	}, nil
}

// ==============================
// Mutations (transaction-aware dengan configs.DB.Begin())
// ==============================

func (service *ItemHistoryService) CreateItemHistory(req *models.ItemHistoryCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.ItemHistory, error) {
	_ = ctx

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	changeType := strings.ToLower(req.ChangeType)

	// pastikan item ada (di dalam tx)
	item, err := service.ItemRepository.FindById(tx, req.ItemID.String(), false)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if item == nil {
		tx.Rollback()
		return nil, fmt.Errorf("item not found")
	}

	// ambil last state (di dalam tx)
	lastHistory, _ := service.ItemHistoryRepository.FindLastByItem(tx, req.ItemID, changeType)

	var oldPrice, oldStock int
	if lastHistory != nil {
		oldPrice = lastHistory.CurrentPrice
		oldStock = lastHistory.CurrentStock
	} else {
		// fallback dari item bila belum ada history
		oldPrice = item.Price
		oldStock = item.Stock
	}

	newH := &models.ItemHistory{
		ID:           uuid.New(),
		ItemID:       req.ItemID,
		ChangeType:   changeType,
		Description:  req.Description,
		CreatedBy:    &userInfo.ID,
		UpdatedBy:    &userInfo.ID,
		OldPrice:     oldPrice,
		OldStock:     oldStock,
		CurrentPrice: oldPrice,
		CurrentStock: oldStock,
	}

	switch changeType {
	case "create_price", "update_price":
		newH.NewPrice = req.NewPrice
		newH.CurrentPrice = req.NewPrice
		item.Price = req.NewPrice
	case "create_stock", "update_stock":
		newH.NewStock = req.NewStock
		newH.CurrentStock = req.NewStock
		item.Stock = req.NewStock
	default:
		tx.Rollback()
		return nil, fmt.Errorf("invalid change_type '%s'", changeType)
	}

	// insert history (tx-aware)
	created, err := service.ItemHistoryRepository.Insert(tx, newH)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// update item (tx-aware)
	if _, err := service.ItemRepository.Update(tx, item); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return created, nil
}

func (service *ItemHistoryService) DeleteItemHistories(req *models.ItemHistoryIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	if len(req.IDs) == 0 {
		return errors.New("itemHistoryIds cannot be empty")
	}

	for _, histID := range req.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for itemHistory delete %v: %v\n", histID, tx.Error)
			return errors.New("error beginning transaction")
		}
		defer func(tx *gorm.DB) {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}(tx)

		if _, err := service.ItemHistoryRepository.FindById(tx, histID.String(), false); err != nil {
			tx.Rollback()
			if err == repositories.ErrItemHistoryNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("ItemHistory not found: %v\n", histID)
				continue
			}
			log.Printf("Error finding itemHistory %v: %v\n", histID, err)
			return errors.New("error finding itemHistory")
		}

		isHard := req.IsHardDelete == "hardDelete"
		if err := service.ItemHistoryRepository.Delete(tx, histID.String(), isHard); err != nil {
			tx.Rollback()
			if isHard {
				log.Printf("Error hard deleting itemHistory %v: %v\n", histID, err)
				return errors.New("error hard deleting itemHistory")
			}
			log.Printf("Error soft deleting itemHistory %v: %v\n", histID, err)
			return errors.New("error soft deleting itemHistory")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing delete for itemHistory %v: %v\n", histID, err)
			return errors.New("error committing itemHistory delete")
		}
	}

	return nil
}

func (service *ItemHistoryService) RestoreItemHistories(req *models.ItemHistoryRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.ItemHistory, error) {
	var restored []models.ItemHistory
	if len(req.IDs) == 0 {
		return nil, errors.New("itemHistoryIds cannot be empty")
	}

	for _, histID := range req.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for itemHistory restore %v: %v\n", histID, tx.Error)
			return nil, errors.New("error beginning transaction")
		}
		defer func(tx *gorm.DB) {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}(tx)

		if _, err := service.ItemHistoryRepository.FindById(tx, histID.String(), true); err != nil {
			tx.Rollback()
			if err == repositories.ErrItemHistoryNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("ItemHistory not found for restore: %v\n", histID)
				continue
			}
			log.Printf("Error finding itemHistory %v: %v\n", histID, err)
			return nil, errors.New("error finding itemHistory")
		}

		got, err := service.ItemHistoryRepository.Restore(tx, histID.String())
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrItemHistoryNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("ItemHistory not found for restore: %v\n", histID)
				continue
			}
			log.Printf("Error restoring itemHistory %v: %v\n", histID, err)
			return nil, errors.New("error restoring itemHistory")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing itemHistory restore %v: %v\n", histID, err)
			return nil, errors.New("error committing itemHistory restore")
		}

		restored = append(restored, *got)
	}

	return restored, nil
}
