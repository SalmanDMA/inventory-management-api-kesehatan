package services

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ItemService struct {
	UploadRepository repositories.UploadRepository
	ItemRepository repositories.ItemRepository
	ItemHistoryRepository repositories.ItemHistoryRepository
}

func NewItemService(uploadRepo repositories.UploadRepository, itemRepo repositories.ItemRepository, itemHistoryRepo repositories.ItemHistoryRepository) *ItemService {
	return &ItemService{
		UploadRepository: uploadRepo,
		ItemRepository: itemRepo,
		ItemHistoryRepository: itemHistoryRepo,
	}
}

func (service *ItemService) GetAllItems() ([]models.ResponseGetItem, error) {
	items, err := service.ItemRepository.FindAll()
	if err != nil {
		return nil, err
	}

	var itemsResponse []models.ResponseGetItem
	for _, item := range items {
		itemsResponse = append(itemsResponse, models.ResponseGetItem{
			ID:          item.ID,
			Name:        item.Name,
			Code:        item.Code,
			Price:       item.Price,
			ImageID:     item.ImageID,
			Image:       item.Image,
			CategoryID:  item.CategoryID,
			Category:    item.Category,
			Description: item.Description,
			Stock:  item.Stock,
			LowStock: item.LowStock,
			ItemHistories: item.ItemHistories,
			CreatedAt: 	item.CreatedAt,
			UpdatedAt: 	item.UpdatedAt,
			DeletedAt: 	item.DeletedAt,
		})
	}

	return itemsResponse, nil
}

func (service *ItemService) GetAllItemsPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.ItemPaginatedResponse, error) {
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

	items, totalCount, err := service.ItemRepository.FindAllPaginated(req)
	if err != nil {
		return nil, err
	}

	itemsResponse := []models.ResponseGetItem{}
	for _, item := range items {
		itemsResponse = append(itemsResponse, models.ResponseGetItem{
			ID:          item.ID,
			Name:        item.Name,
			Code:        item.Code,
			Price:       item.Price,
			ImageID:     item.ImageID,
			Image:       item.Image,
			CategoryID:  item.CategoryID,
			Category:    item.Category,
			Description: item.Description,
			Stock:  item.Stock,
			LowStock: item.LowStock,
			ItemHistories: item.ItemHistories,
			CreatedAt: 	item.CreatedAt,
			UpdatedAt: 	item.UpdatedAt,
			DeletedAt: 	item.DeletedAt,
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

	return &models.ItemPaginatedResponse{
		Data:       itemsResponse,
		Pagination: paginationResponse,
	}, nil
}

func (service *ItemService) GetItemByID(itemId string) (*models.ResponseGetItem, error) {
	item, err := service.ItemRepository.FindById(itemId, false)

	if err != nil {
		return nil, err
	}

	return &models.ResponseGetItem{
		 ID:          item.ID,
			Name:        item.Name,
			Code:        item.Code,
			Price:       item.Price,
			ImageID:     item.ImageID,
			Image:       item.Image,
			CategoryID:  item.CategoryID,
			Category:    item.Category,
			Description: item.Description,
			Stock:  					item.Stock,
			LowStock: item.LowStock,
			ItemHistories: item.ItemHistories,
	}, nil
}

func (service *ItemService) CreateItem(itemRequest *models.ItemCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Item, error) {
	if _, err := service.ItemRepository.FindByName(itemRequest.Name); err == nil {
		return nil, errors.New("item already exists") 
	} else if err != repositories.ErrItemNotFound {
		 return nil, errors.New("error checking item: " + err.Error())
	}

	newItem := &models.Item{
		ID:          uuid.New(),
		Name:        itemRequest.Name,
		Code:        itemRequest.Code,
		Price:       itemRequest.Price,
		Stock:       itemRequest.Stock,
		LowStock:    itemRequest.LowStock,
		CategoryID:  itemRequest.CategoryID,
		Description: itemRequest.Description,
	}

	var imageUUIDStr string
	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			if imageUUIDStr != "" {
				helpers.DeleteLocalFileImmediate(imageUUIDStr)
			}
		}
	}()

	file, err := ctx.FormFile("image")
	if err == nil && file != nil {
		imageUUIDStr, err = helpers.SaveFile(ctx, file, "items")
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to save image: %w", err)
		}

		imageUUID, err := uuid.Parse(imageUUIDStr)
		if err != nil {
			tx.Rollback()
			helpers.DeleteLocalFileImmediate(imageUUIDStr)
			return nil, fmt.Errorf("invalid image UUID: %w", err)
		}
		newItem.ImageID = &imageUUID
	}

	result := tx.Create(newItem)
	if result.Error != nil {
		tx.Rollback()
		if imageUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(imageUUIDStr)
		}
		return nil, fmt.Errorf("error creating item: %w", result.Error)
	}

	if err := tx.Commit().Error; err != nil {
		if imageUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(imageUUIDStr)
		}
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	insertedItem, err := service.ItemRepository.FindById(newItem.ID.String(), false)
	if err != nil {
		log.Printf("Warning: Item created but failed to fetch created data: %v", err)
		return newItem, nil
	}

	if _, err := service.ItemHistoryRepository.Insert(&models.ItemHistory{
		ID:          uuid.New(),
		ItemID:      insertedItem.ID,
		ChangeType:  "create_price",
		OldPrice:    0,
		NewPrice:    insertedItem.Price,
		CurrentPrice: insertedItem.Price,
		CreatedBy:   &userInfo.ID,
		UpdatedBy:   &userInfo.ID,
		Description: "Initial price set to " + strconv.Itoa(insertedItem.Price),
	}); err != nil {
		return nil, err
	}

	if _, err := service.ItemHistoryRepository.Insert(&models.ItemHistory{
		ID:          uuid.New(),
		ItemID:      insertedItem.ID,
		ChangeType:  "create_stock",
		OldStock:    0,
		NewStock:    insertedItem.Stock,
		CurrentStock: insertedItem.Stock,
		CreatedBy:   &userInfo.ID,
		UpdatedBy:   &userInfo.ID,
		Description: "Initial stock set to " + strconv.Itoa(insertedItem.Stock),
	}); err != nil {
		return nil, err
	}

	return insertedItem, nil
}

func (service *ItemService) UpdateItem(itemRequest *models.ItemUpdateRequest, itemID string, ctx *fiber.Ctx, userInfo *models.User) (*models.Item, error) {
	item, err := service.ItemRepository.FindById(itemID, false)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("item not found")
	}

	if itemRequest.Name != "" {
		item.Name = itemRequest.Name
	}
	if itemRequest.Code != "" {
		item.Code = itemRequest.Code
	}
	if itemRequest.LowStock != 0 {
		item.LowStock = itemRequest.LowStock
	}
	if itemRequest.CategoryID != uuid.Nil {
		item.CategoryID = itemRequest.CategoryID
	}
	if itemRequest.Description != "" {
		item.Description = itemRequest.Description
	}

	file, err := ctx.FormFile("image")
	if err == nil && file != nil {
		oldImageID := item.ImageID
		var newImageUUIDStr string
		var newImageID uuid.UUID

		tx := configs.DB.Begin()
		if tx.Error != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
		}

		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				if newImageUUIDStr != "" {
					helpers.DeleteLocalFileImmediate(newImageUUIDStr)
				}
			}
		}()

		newImageUUIDStr, err := helpers.SaveFile(ctx, file, "items")
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to save image: %w", err)
		}

		newImageID, err = uuid.Parse(newImageUUIDStr)
		if err != nil {
			tx.Rollback()
			helpers.DeleteLocalFileImmediate(newImageUUIDStr)
			return nil, fmt.Errorf("invalid image UUID: %w", err)
		}

		item.Image = nil
		item.ImageID = &newImageID
		result := tx.Model(item).Select("*").Updates(item)
		if result.Error != nil {
			tx.Rollback()
			helpers.DeleteLocalFileImmediate(newImageUUIDStr)
			return nil, fmt.Errorf("error updating item with image: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			tx.Rollback()
			helpers.DeleteLocalFileImmediate(newImageUUIDStr)
			return nil, errors.New("no rows affected, item may not exist")
		}

		if err := tx.Commit().Error; err != nil {
			helpers.DeleteLocalFileImmediate(newImageUUIDStr)
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		updatedItem, err := service.ItemRepository.FindById(item.ID.String(), false)
		if err != nil {
			log.Printf("Warning: Item updated but failed to fetch updated data: %v", err)
			return item, nil
		}

		if oldImageID != nil {
			go func() {
				if err := helpers.DeleteLocalFileImmediate(oldImageID.String()); err != nil {
					log.Printf("Warning: Failed to delete old image file %s: %v", oldImageID.String(), err)
				}
			}()
		}

		return updatedItem, nil
	}

	updateItem, err := service.ItemRepository.Update(item)
	if err != nil {
		return nil, err
	}

	return updateItem, nil
}

func (service *ItemService) DeleteItems(itemRequest *models.ItemIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, itemId := range itemRequest.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for item %v: %v\n", itemId, tx.Error)
			return errors.New("error beginning transaction")
		}

		item, err := service.ItemRepository.FindById(itemId.String(), true)
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrItemNotFound {
				log.Printf("Item not found: %v\n", itemId)
				continue
			}
			log.Printf("Error finding item %v: %v\n", itemId, err)
			return errors.New("error finding item")
		}

		if itemRequest.IsHardDelete == "hardDelete" {
			if item.ImageID != nil && item.ImageID.String() != "" {
				if err := tx.Unscoped().Delete(&models.Item{}, "id = ?", itemId).Error; err != nil {
					tx.Rollback()
					log.Printf("Error hard deleting item %v: %v\n", itemId, err)
					return errors.New("error hard deleting item")
				}

				if err := tx.Commit().Error; err != nil {
					log.Printf("Error committing hard delete for item %v: %v\n", itemId, err)
					return errors.New("error committing hard delete")
				}

				go func(imageID string) {
					if err := helpers.DeleteLocalFileImmediate(imageID); err != nil {
						log.Printf("Warning: Failed to delete image file %s: %v", imageID, err)
					}
				}(item.ImageID.String())
			} else {
				if err := tx.Unscoped().Delete(&models.Item{}, "id = ?", itemId).Error; err != nil {
					tx.Rollback()
					log.Printf("Error hard deleting item %v: %v\n", itemId, err)
					return errors.New("error hard deleting item")
				}

				if err := tx.Commit().Error; err != nil {
					log.Printf("Error committing hard delete for item %v: %v\n", itemId, err)
					return errors.New("error committing hard delete")
				}
			}
		} else {
			if err := tx.Delete(&models.Item{}, "id = ?", itemId).Error; err != nil {
				tx.Rollback()
				log.Printf("Error soft deleting item %v: %v\n", itemId, err)
				return errors.New("error soft deleting item")
			}

			if err := tx.Commit().Error; err != nil {
				log.Printf("Error committing soft delete for item %v: %v\n", itemId, err)
				return errors.New("error committing soft delete")
			}
		}
	}

	return nil
}

func (service *ItemService) RestoreItems(itemRequest *models.ItemRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Item, error) {
	var restoredItems []models.Item

	for _, itemId := range itemRequest.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for item restore %v: %v\n", itemId, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		result := tx.Model(&models.Item{}).Unscoped().Where("id = ?", itemId).Update("deleted_at", nil)
		if result.Error != nil {
			tx.Rollback()
			log.Printf("Error restoring item %v: %v\n", itemId, result.Error)
			return nil, errors.New("error restoring item")
		}

		if result.RowsAffected == 0 {
			tx.Rollback()
			log.Printf("Item not found for restore: %v\n", itemId)
			continue
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing item restore %v: %v\n", itemId, err)
			return nil, errors.New("error committing item restore")
		}

		restoredItem, err := service.ItemRepository.FindById(itemId.String(), true)
		if err != nil {
			log.Printf("Error fetching restored item %v: %v\n", itemId, err)
			continue
		}

		restoredItems = append(restoredItems, *restoredItem)
	}

	return restoredItems, nil
}